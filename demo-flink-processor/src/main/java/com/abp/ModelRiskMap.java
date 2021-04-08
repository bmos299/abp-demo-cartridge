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

import org.apache.flink.api.common.functions.MapFunction;

import com.abp.rest.client.ModelInferResponse;
import com.abp.rest.client.ModelRestClient;

/**
 * This is Model risk map which makes call to the model serving with pay delay
 * and do the mapping based on the output received
 * 
 * @author amarpandey
 *
 */
public class ModelRiskMap implements MapFunction<Invoice, Invoice> {

    private static final long serialVersionUID = 1L;
    public static final String RISK_LOW = "Low";
    public static final String RISK_MEDIUM = "Medium";
    public static final String RISK_HIGH = "High";
    public static final String DEFAULT_AI_MODEL_NAME = "anomaly-classifier-predictor";

    private String predictorUrl = null;

    public ModelRiskMap(String predictorUrl) {
        this.predictorUrl = predictorUrl;
    }

    @Override
    public Invoice map(Invoice inv) throws Exception {
        ModelRestClient modelRestClient = new ModelRestClient(predictorUrl);
        ModelInferResponse response = modelRestClient.callRestInferenceService(inv.Pay_Delay);
        int out = response.getOutput();
        inv.Risk = RISK_LOW;
        if (out > 50)
            inv.Risk = RISK_MEDIUM;
        if (out > 100)
            inv.Risk = RISK_HIGH;
        if (response.getModelName() == null)
            inv.Model_Name = DEFAULT_AI_MODEL_NAME;
        else
            inv.Model_Name = response.getModelName();
        return inv;
    }
}
