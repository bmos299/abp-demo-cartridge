/********************************************************** {COPYRIGHT-TOP} ****
 * Licensed Materials - Property of IBM
 * 5900-AEO
 *
 * Copyright IBM Corp. 2020, 2021. All Rights Reserved.
 *
 * US Government Users Restricted Rights - Use, duplication, or
 * disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
 ********************************************************** {COPYRIGHT-END} ***/
package com.abp.grpc.client;

import static io.grpc.Metadata.ASCII_STRING_MARSHALLER;

import java.util.HashSet;
import java.util.Random;
import java.util.Set;

import com.google.protobuf.ByteString;

import inference.GRPCInferenceServiceGrpc;
import inference.GrpcService.InferTensorContents;
import inference.GrpcService.ModelInferRequest;
import inference.GrpcService.ModelInferRequest.InferInputTensor;
import inference.GrpcService.ModelInferResponse;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.Metadata;
import io.grpc.stub.MetadataUtils;

/**
 * Grpc client to send grpc request to models
 * @author amarpandey
 *
 */

public class GrpcClient {
   private GRPCInferenceServiceGrpc.GRPCInferenceServiceBlockingStub stub = null;
   private Metadata metadata = null;
   private ManagedChannel channel = null;

    public GrpcClient(String predictorUrl) {
        this.metadata = new Metadata();
        final Metadata.Key<String> MODEL_METADATA_KEY = Metadata.Key.of("mm-vmodel-id", ASCII_STRING_MARSHALLER);
        metadata.put(MODEL_METADATA_KEY, "anomaly-classifier-predictor");
        this.channel = ManagedChannelBuilder.forTarget(predictorUrl).usePlaintext().build();
        this.stub = MetadataUtils
                .attachHeaders(GRPCInferenceServiceGrpc.newBlockingStub(channel), metadata);
    }

    public ModelInferResponse callGrpcInferenceService(int payDelay) {
        ModelInferResponse modelInferResponse = this.stub.modelInfer(ModelInferRequest.newBuilder()
                .setModelName("anomaly-classifier-predictor")
                .addInputs(InferInputTensor.newBuilder().setName("Pay_Delay").addShape(1).addShape(1)
                        .setDatatype("INT64").setContents(InferTensorContents.newBuilder().addInt64Contents(payDelay)))
                .build());

        return modelInferResponse;
    }

    public void shutdown() {
        this.channel.shutdown();
    }

    public static void main(String[] args) throws InterruptedException {

        Set<Integer> numbers = new HashSet<Integer>();

        for (int i = 0; i < 500; i++) {
            numbers.add(getRandomNumberInRange(1, 500));
        }

        Metadata metadata = new Metadata();
        final Metadata.Key<String> MODEL_METADATA_KEY = Metadata.Key.of("mm-vmodel-id", ASCII_STRING_MARSHALLER);
        metadata.put(MODEL_METADATA_KEY, "anomaly-classifier-predictor");

        ManagedChannel channel = ManagedChannelBuilder.forAddress("indigo", 8033).usePlaintext().build();

        GRPCInferenceServiceGrpc.GRPCInferenceServiceBlockingStub stub = MetadataUtils
                .attachHeaders(GRPCInferenceServiceGrpc.newBlockingStub(channel), metadata);
        System.out.println("Value\tRisk");
        for (int number : numbers) {
            ModelInferResponse modelInferResponse = stub
                    .modelInfer(ModelInferRequest.newBuilder().setModelName("anomaly-classifier-predictor")
                            .addInputs(InferInputTensor.newBuilder().setName("Pay_Delay").addShape(1).addShape(1)
                                    .setDatatype("INT64")
                                    .setContents(InferTensorContents.newBuilder().addInt64Contents(number)))
                            .build());

            ByteString bs = modelInferResponse.getRawOutputContentsList().get(0);
            byte[] byteArray = bs.toByteArray();
            Byte b = byteArray[0];
            int out = b.intValue();
            String risk = "Low";
            if (out > 50)
                risk = "Medium";
            if (out > 100)
                risk = "High";

            System.out.println(number + "\t" + risk);
        }

        channel.shutdown();
    }

    private static int getRandomNumberInRange(int min, int max) {
        Random r = new Random();
        return r.nextInt((max - min) + 1) + min;
    }

}
