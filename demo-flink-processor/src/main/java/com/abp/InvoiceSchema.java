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

import java.io.IOException;
import java.net.URI;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

import org.apache.flink.api.common.serialization.DeserializationSchema;
import org.apache.flink.api.common.serialization.SerializationSchema;
import org.apache.flink.api.common.typeinfo.TypeInformation;
import org.apache.flink.api.java.typeutils.TypeExtractor;

import io.cloudevents.CloudEvent;
import io.cloudevents.core.builder.CloudEventBuilder;
import io.cloudevents.core.provider.EventFormatProvider;
import io.cloudevents.jackson.JsonFormat;

// Custom schema to help with serialization / deserialization
public class InvoiceSchema implements DeserializationSchema<Invoice>, SerializationSchema<Invoice> {

    private static final long serialVersionUID = 1L;
    ObjectMapper mapper = new ObjectMapper();

    @Override
    public Invoice deserialize(byte[] bytes) throws IOException {
        Invoice invoice = mapper.readValue(new String(bytes), Invoice.class);

        return invoice;
    }

    @Override
    public byte[] serialize(Invoice invoice) {
        String invoiceJson = "";
        try {
            invoiceJson = mapper.writeValueAsString(invoice);
        } catch (JsonProcessingException e) {
            e.printStackTrace();
        }

        CloudEvent event = CloudEventBuilder.v1().withData(invoiceJson.getBytes()).withId(invoice.Invoice_ID).withType("com.abp.risk").withDataContentType("application/json")
                .withSource(URI.create("https://github.ibm.com/automation-base-pak/abp-demo-cartridge")).build();

        return EventFormatProvider.getInstance().resolveFormat(JsonFormat.CONTENT_TYPE).serialize(event);
    }

    @Override
    public TypeInformation<Invoice> getProducedType() {
        return TypeExtractor.getForClass(Invoice.class);
    }

    @Override
    public boolean isEndOfStream(Invoice nextElement) {
        return false;
    }
}
