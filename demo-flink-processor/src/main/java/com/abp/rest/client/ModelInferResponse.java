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

public class ModelInferResponse {
    private String modelName;
    private String modelVersion;
    private int id;
    private int output;
    
    public String getModelName() {
        return modelName;
    }
    public void setModelName(String modelName) {
        this.modelName = modelName;
    }
    public String getModelVersion() {
        return modelVersion;
    }
    public void setModelVersion(String modelVersion) {
        this.modelVersion = modelVersion;
    }
    public int getId() {
        return id;
    }
    public void setId(int id) {
        this.id = id;
    }
    public int getOutput() {
        return output;
    }
    public void setOutput(int output) {
        this.output = output;
    }
}
