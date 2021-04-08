/********************************************************** {COPYRIGHT-TOP} ****
 * Licensed Materials - Property of IBM
 * 5900-AEO
 *
 * Copyright IBM Corp. 2021. All Rights Reserved.
 *
 * US Government Users Restricted Rights - Use, duplication, or
 * disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
 ********************************************************** {COPYRIGHT-END} ***/
package com.abp.rest.client;

import java.io.BufferedReader;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.math.RoundingMode;
import java.nio.charset.Charset;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.security.KeyStore;

import javax.net.ssl.SSLContext;

import org.apache.commons.io.Charsets;
import org.apache.http.Header;
import org.apache.http.HttpEntity;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClientBuilder;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.ssl.SSLContextBuilder;
import org.apache.http.ssl.SSLContexts;
import org.apache.http.util.EntityUtils;
import org.json.JSONArray;
import org.json.JSONObject;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class ModelRestClient {

    private CloseableHttpClient closeableHttpClient;
    private String predictorUrl;

    Logger LOGGER = LoggerFactory.getLogger(ModelRestClient.class);
    
    public ModelRestClient(String predictorUrl) {
        LOGGER.info("Creating REST model client");
        BufferedReader bufferedReader = null;
        this.predictorUrl = predictorUrl;

        String aiModelTlsVersion = System.getenv("AI_MODEL_TLS_VERSION");
        if (aiModelTlsVersion != null) {
            String aiModelTrustStorePath = System.getenv("TRUSTSTORE_PATH");
            String truststorePasswordFile = System.getenv("TRUSTSTORE_PASSWORD_PATH");

            if (aiModelTrustStorePath != null && !aiModelTrustStorePath.isEmpty() && truststorePasswordFile != null) {
                try {
                    bufferedReader = new BufferedReader(
                            new InputStreamReader(new FileInputStream(truststorePasswordFile), StandardCharsets.UTF_8));
                    String keyStorePass = bufferedReader.readLine();
                    Path trustStorePath = Paths.get(aiModelTrustStorePath);
                    String trustStoreType = System.getenv().getOrDefault("TRUSTSTORE_TYPE", "PKCS12");
                    KeyStore aiModelTruststore = KeyStore.getInstance(trustStoreType);
                    InputStream is = Files.newInputStream(trustStorePath);
                    aiModelTruststore.load(is, keyStorePass == null ? "".toCharArray() : keyStorePass.toCharArray());
                    SSLContextBuilder sslBuilder = SSLContexts.custom().loadTrustMaterial(aiModelTruststore, null);
                    SSLContext sslContext = sslBuilder.build();
                    HttpClientBuilder clientbuilder = HttpClients.custom();

                    clientbuilder.setSSLContext(sslContext);

                    CloseableHttpClient httpClient = clientbuilder.build();

                    this.closeableHttpClient = httpClient;
                } catch (Exception exception) {
                    System.err.println("Failed to read ai model truststore, Exception: " + exception);
                } finally {
                    try {
                        bufferedReader.close();
                    } catch (IOException e) {
                        System.err.println("Failed to close bufferedReader, Exception: " + e);
                    }
                }
            }
        } else {
            this.closeableHttpClient = HttpClientBuilder.create().build();
        }
    }

    /**
     * Method to call rest based inference service e.g. kfserving
     * 
     * @param predictorUrl
     * @param inputDataFile
     * @return
     * @throws Exception
     */
    public ModelInferResponse callRestInferenceService(int payDelay) throws Exception {
        ModelInferResponse result = null;
        String predictorUrlWithPath = this.predictorUrl;
        HttpPost httpPost = new HttpPost(predictorUrlWithPath);
        String inputJson = "{\"instances\":[{\"Pay_Delay\":" + payDelay + "}]}";

        httpPost.setHeader("Accept", "application/json");
        httpPost.setHeader("Content-type", "application/json");
        StringEntity stringEntity = new StringEntity(inputJson);
        httpPost.setEntity(stringEntity);
        CloseableHttpResponse response = this.closeableHttpClient.execute(httpPost);
        try {
            LOGGER.info("    Response status " + response.getStatusLine());
            HttpEntity entity = response.getEntity();
            Header encodingHeader = entity.getContentEncoding();

            Charset encoding = encodingHeader == null ? StandardCharsets.UTF_8
                    : Charsets.toCharset(encodingHeader.getValue());

            String json = EntityUtils.toString(entity, encoding);
            JSONObject o = new JSONObject(json);
            LOGGER.info("Result from the inference service is:  " + o.toString());
            JSONArray jsonArray = o.getJSONArray("predictions");
            if(!jsonArray.isEmpty())
            {
                JSONArray arr = jsonArray.getJSONArray(0);

                result = new ModelInferResponse();
                result.setOutput(arr.getBigDecimal(0).setScale(0, RoundingMode.UP).intValue());
                
                //setting default version 1, TODO need to check as we don't get this in response.
                result.setModelVersion("1");
            }
            EntityUtils.consume(entity);

        } finally {
            response.close();
        }

        return result;
    }

    /**
     * Utility method
     * 
     * @param file
     * @return
     * @throws Exception
     */
    public static String readFileAsString(String file) throws Exception {
        return new String(Files.readAllBytes(Paths.get(file)));
    }
}
