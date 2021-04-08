/********************************************************** {COPYRIGHT-TOP} ****
 * Licensed Materials - Property of IBM
 * 5900-AEO
 *
 * Copyright IBM Corp. 2020, 2021. All Rights Reserved.
 *
 * US Government Users Restricted Rights - Use, duplication, or
 * disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
 ********************************************************** {COPYRIGHT-END} ***/
package com.abp;

import java.nio.charset.StandardCharsets;
import java.util.Properties;

import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;
import org.apache.flink.streaming.connectors.kafka.FlinkKafkaConsumer;
import org.apache.flink.streaming.connectors.kafka.FlinkKafkaProducer;
import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.io.FileInputStream;

import org.apache.flink.api.common.functions.RuntimeContext;
import org.apache.flink.api.java.utils.ParameterTool;

import org.apache.flink.streaming.connectors.elasticsearch.ElasticsearchSinkFunction;
import org.apache.flink.streaming.connectors.elasticsearch.RequestIndexer;
import org.apache.flink.streaming.connectors.elasticsearch7.ElasticsearchSink;

import org.apache.http.HttpHost;
import org.apache.http.auth.AuthScope;
import org.apache.http.auth.UsernamePasswordCredentials;
import org.apache.http.client.CredentialsProvider;
import org.apache.http.impl.client.BasicCredentialsProvider;
import org.apache.http.impl.nio.client.HttpAsyncClientBuilder;
import org.apache.http.message.BasicHeader;

import org.elasticsearch.action.index.IndexRequest;
import org.elasticsearch.client.Requests;

import javax.net.ssl.SSLContext;
import org.apache.http.ssl.SSLContexts;
import org.apache.http.ssl.SSLContextBuilder;
import java.nio.file.Paths;
import java.security.KeyStore;
import java.nio.file.Files;
import java.nio.file.Path;
import java.io.InputStream;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.*;

// Main class for the Flink job
public class Processor {
  private static final String DEFAULT_KAFKA_BOOTSTRAP_SERVERS = "iaf-system-kafka-bootstrap:9092";
  private static final String DEFAULT_GROUP_ID = "iafdemo-flink";
  private static final String DEFAULT_RAW_TOPIC = "iafdemo-raw";
  private static final String DEFAULT_RISK_TOPIC = "iafdemo-anomaly";
  private static final String DEFAULT_ELASTIC_HOST = "https://iaf-system-elasticsearch-es:9200";
  private static final String DEFAULT_ELASTIC_RAW_INDEX = "iafdemo-raw";
  private static final String DEFAULT_ELASTIC_RISK_INDEX = "iafdemo-anomaly";

  private static final String ELASTIC_USERNAME = "elasticsearch-admin";
  private static final String ELASTIC_PASSWORD = "***";

  private static final Properties properties = new Properties();

  public static void main(final String[] args) throws Exception {
    Logger LOGGER = LoggerFactory.getLogger(Processor.class);
    ParameterTool parameter = ParameterTool.fromArgs(args);
    // Read args first and log. Check JobManager for these - they will not 
    // be available in User Configuration for the Flink UI unless the job completes.
    String argsToPrint = Arrays.toString(args);
    LOGGER.info("Args for job: {}", argsToPrint);
    String groupId = parameter.get("groupId");
    if (groupId == null || groupId.equals("")) {
      groupId = DEFAULT_GROUP_ID;
      LOGGER.info("Using default Kafka group ID as none provided through args");
    }

    String rawTopic = parameter.get("rawTopic");
    if (rawTopic == null || rawTopic.equals("")) {
      rawTopic = DEFAULT_RAW_TOPIC;
      LOGGER.info("Using default raw topic name as none provided through args");
    }

    String riskTopic = parameter.get("riskTopic");
    if (riskTopic == null || riskTopic.equals("")) {
      riskTopic = DEFAULT_RISK_TOPIC;
      LOGGER.info("Using default risk topic name as none provided through args");
    }

    final String esHost = System.getenv().getOrDefault("ELASTIC_URI", DEFAULT_ELASTIC_HOST);
    final String bootstrapServers = System.getenv().getOrDefault("KAFKA_BOOTSTRAP_SERVERS", DEFAULT_KAFKA_BOOTSTRAP_SERVERS);
    String predictorUrl = parameter.get("modelPredictorURL");
    if (predictorUrl == null || predictorUrl.equals("")) {
      LOGGER.warn("No model predictor URL found");
    }

    properties.setProperty("bootstrap.servers", bootstrapServers);
    properties.setProperty("group.id", groupId);

    LOGGER.info("Kafka bootstrap servers: " + bootstrapServers);
    LOGGER.info("Kafka group ID: " + groupId);
    LOGGER.info("Raw topic name: " + rawTopic);
    LOGGER.info("Risk topic name: " + riskTopic);
    LOGGER.info("Elastic host name: " + esHost);
    LOGGER.info("Predictor URL: " + predictorUrl);

    // Create execution environment
    StreamExecutionEnvironment env = StreamExecutionEnvironment.getExecutionEnvironment();

    String authtype = System.getenv("KAFKA_AUTH_TYPE");
    Boolean isScramAuth = authtype != null && !authtype.isEmpty() && authtype.equals("SCRAM-SHA-512");
    if (isScramAuth) {
      String username = System.getenv("KAFKA_SCRAM_USERNAME");
      String passwordFile = System.getenv("KAFKA_SCRAM_PASSWORD_PATH");
      if(username != null && !username.isEmpty() && passwordFile != null && !passwordFile.isEmpty()) {
        BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(new FileInputStream(passwordFile), StandardCharsets.UTF_8));
        String password = bufferedReader.readLine();
        properties.setProperty("security.protocol", "SASL_PLAINTEXT");
        properties.setProperty("sasl.mechanism", "SCRAM-SHA-512");
        properties.setProperty("sasl.jaas.config", "org.apache.kafka.common.security.scram.ScramLoginModule required \nusername=\"" + username + "\" \npassword=\"" + password + "\";");
      }
    }

    String kafkaTlsVersion = System.getenv("KAFKA_TLS_VERSION");
    String truststore = System.getenv("KAFKA_TRUSTSTORE_PATH");
		String truststorePasswordFile = System.getenv("KAFKA_TRUSTSTORE_PASSWORD_PATH");
    Boolean isTls = kafkaTlsVersion != null && truststore != null && !truststore.isEmpty() && truststorePasswordFile != null && !truststorePasswordFile.isEmpty();

    if (isTls) {
			BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(new FileInputStream(truststorePasswordFile), StandardCharsets.UTF_8));
			String truststorePassword = bufferedReader.readLine();
			properties.setProperty("ssl.truststore.location", truststore);
			properties.setProperty("ssl.truststore.password", truststorePassword);
      String truststoreType = System.getenv().getOrDefault("KAFKA_TRUSTSTORE_TYPE", "PKCS12");
			properties.setProperty("ssl.truststore.type", truststoreType);
      properties.setProperty("security.protocol", "SSL");
    }

    if (isScramAuth && isTls) {
      properties.setProperty("security.protocol", "SASL_SSL");
    }

		// Setup source for raw events
		FlinkKafkaConsumer<RawInput> kafkaConsumer = new FlinkKafkaConsumer<>(rawTopic, new RawInputSchema(), properties);
		kafkaConsumer.setStartFromEarliest();

		// Read valid invoices
		DataStream<RawInput> rawInputStream = env.addSource(kafkaConsumer);

    List<HttpHost> esHttphost = new ArrayList<>();
    esHttphost.add(HttpHost.create(esHost));

    // use a ElasticsearchSink.Builder to create an ElasticsearchSink
    // use a ElasticsearchSink.Builder to create an ElasticsearchSink
    ElasticsearchSink.Builder<RawInput> esSinkBuilder = new ElasticsearchSink.Builder<>(
      esHttphost,
      new ElasticsearchSinkFunction<RawInput>() {
        private IndexRequest createIndexRequest(RawInput element) {
        // Use here due to the fact local variables referenced from an inner class must be final or effectively final
        String esRawIndex = parameter.get("esRawIndex");
        if (esRawIndex == null || esRawIndex.equals("")) {
          esRawIndex = DEFAULT_ELASTIC_RAW_INDEX;
          LOGGER.info("Using default elastic raw index as none provided through args");
        }
          return Requests.indexRequest()
              .index(esRawIndex)
              .source(element.rawInputToJson());
        }

        @Override
        public void process(RawInput element, RuntimeContext ctx, RequestIndexer indexer) {
          indexer.add(createIndexRequest(element));
        }
      }
    );

    // configuration for the bulk requests; this instructs the sink to emit after every element,
    // otherwise they would be buffered
		// esSinkBuilder.setBulkFlushMaxActions(1);
		esSinkBuilder.setBulkFlushMaxActions(3000);
		esSinkBuilder.setBulkFlushMaxSizeMb(50);
		esSinkBuilder.setBulkFlushInterval(3000);

    // provide a RestClientFactory for custom configuration on the internally created REST client
    esSinkBuilder.setRestClientFactory(restClientBuilder -> {
        restClientBuilder.setDefaultHeaders(new BasicHeader[]{new BasicHeader("Content-Type","application/json")});
        restClientBuilder.setHttpClientConfigCallback(Processor::setElasticCredentials);
    });

    rawInputStream.addSink(esSinkBuilder.build());

		// Read valid invoices
    // Setup source for Invoice events
 		FlinkKafkaConsumer<Invoice> kafkaConsumerInvoice = new FlinkKafkaConsumer<>(rawTopic, new InvoiceSchema(), properties);
 		kafkaConsumerInvoice.setStartFromEarliest();

		DataStream<Invoice> invoiceStream = env.addSource(kafkaConsumerInvoice).filter(new ValidFilter());

		// transform late invoices to risks
    DataStream<Invoice> riskStream;
    // check if the predictorUrl env variable in the event processing task is null or empty,
    // then use the existing map
    if(predictorUrl == null || predictorUrl.equals(""))
        riskStream = invoiceStream.filter(new LateFilter()).map(new RiskMap());
    else
        riskStream = invoiceStream.filter(new LateFilter()).map(new ModelRiskMap(predictorUrl));

		// Setup target for Invoice risk
		FlinkKafkaProducer<Invoice> riskProducer = new FlinkKafkaProducer<>(riskTopic, new InvoiceSchema(), properties);
		riskStream.addSink(riskProducer);

    // use a ElasticsearchSink.Builder to create an ElasticsearchSink
    ElasticsearchSink.Builder<Invoice> esSinkBuilderForAnomaly = new ElasticsearchSink.Builder<>(
      esHttphost,
      new ElasticsearchSinkFunction<Invoice>() {
        private IndexRequest createIndexRequest(Invoice invoice) {
        String esRiskIndex = parameter.get("esRiskIndex");
        if (esRiskIndex == null || esRiskIndex.equals("")) {
          esRiskIndex = DEFAULT_ELASTIC_RISK_INDEX;
          LOGGER.info("Using default elastic risk index as none provided through args");
        }
          return Requests.indexRequest()
              .index(esRiskIndex)
              .source(invoice.toJson());
        }

        @Override
        public void process(Invoice element, RuntimeContext ctx, RequestIndexer indexer) {
          indexer.add(createIndexRequest(element));
        }
      }
    );

    // configuration for the bulk requests; this instructs the sink to emit after every element,
    // otherwise they would be buffered
    // esSinkBuilderForAnomaly.setBulkFlushMaxActions(1);
    esSinkBuilderForAnomaly.setBulkFlushMaxActions(3000);
    esSinkBuilderForAnomaly.setBulkFlushMaxSizeMb(50);
    esSinkBuilderForAnomaly.setBulkFlushInterval(3000);

    // provide a RestClientFactory for custom configuration on the internally created REST client
    esSinkBuilderForAnomaly.setRestClientFactory(restClientBuilder -> {
        restClientBuilder.setDefaultHeaders(new BasicHeader[]{new BasicHeader("Content-Type","application/json")});
        restClientBuilder.setHttpClientConfigCallback(Processor::setElasticCredentials);
    });
    
    riskStream.addSink(esSinkBuilderForAnomaly.build());

    env.execute();
	}


  private static HttpAsyncClientBuilder setElasticCredentials(HttpAsyncClientBuilder httpClientBuilder) {
    String authtype = System.getenv("ELASTIC_AUTH_TYPE");
    if (authtype != null && authtype.equals("BASIC")) {
      String esUser = ELASTIC_USERNAME ;
      String esPwd = ELASTIC_PASSWORD ;

      try {
        String usernamefile = System.getenv("ELASTIC_BASIC_USERNAME_PATH");
        BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(new FileInputStream(usernamefile), StandardCharsets.UTF_8));
        String username = bufferedReader.readLine();
        if (username != null) {
          esUser = username;
        }
      } catch (Exception exception) {
        System.err.println("Failed to read elastic username");
      }

      try {
        String passwordfile = System.getenv("ELASTIC_BASIC_PASSWORD_PATH");
        BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(new FileInputStream(passwordfile), StandardCharsets.UTF_8));
        String password = bufferedReader.readLine();
        if (password != null) {
          esPwd = password;
        }
      } catch (Exception exception) {
        System.err.println("Failed to read elastic password");
      }

      // elastic search basic auth
      CredentialsProvider credentialsProvider = new BasicCredentialsProvider();
      credentialsProvider.setCredentials(AuthScope.ANY, new UsernamePasswordCredentials(esUser, esPwd));
      httpClientBuilder.setDefaultCredentialsProvider(credentialsProvider);
    }

    // Presence of this env means tls is enabled
    String elasticTlsVersion = System.getenv("ELASTIC_TLS_VERSION");
    if (elasticTlsVersion != null) {
      String elasticTruststorePath = System.getenv("ELASTIC_TRUSTSTORE_PATH");
      String truststorePasswordFile = System.getenv("ELASTIC_TRUSTSTORE_PASSWORD_PATH");
      if (elasticTruststorePath != null && !elasticTruststorePath.isEmpty() && truststorePasswordFile != null) {
        try {
          BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(new FileInputStream(truststorePasswordFile), StandardCharsets.UTF_8));
          String keyStorePass = bufferedReader.readLine();

          Path trustStorePath = Paths.get(elasticTruststorePath);
          String elasticType = System.getenv().getOrDefault("ELASTIC_TRUSTSTORE_TYPE", "PKCS12");
          KeyStore elasticTruststore = KeyStore.getInstance(elasticType);
          InputStream is = Files.newInputStream(trustStorePath);
          elasticTruststore.load(is, keyStorePass.toCharArray());
          SSLContextBuilder sslBuilder = SSLContexts.custom().loadTrustMaterial(elasticTruststore, null);
          SSLContext sslContext = sslBuilder.build();
          httpClientBuilder.setSSLContext(sslContext);
        } catch (Exception exception) {
          System.err.println("Cannot extract elastic truststore, Exception: " + exception);
        }
      }
    }

    return httpClientBuilder;
  }
}
