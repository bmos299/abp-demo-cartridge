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

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import java.util.*;

// POJO for storing  event information
@JsonIgnoreProperties(ignoreUnknown = true)
public class Invoice {
    public String Invoice_ID;
    public float Invoice_Amount;
    public String Invoice_Due_Date;
    public String Pay_Type;
    public int Pay_Delay;
    public String Risk;

    //add model predictor name
    public String Model_Name;

    public Invoice() {
    }

    public Map<String, Object> toJson() {
        HashMap<String, Object> json = new HashMap<>();
        json.put("Invoice_ID", Invoice_ID);
        json.put("Invoice_Amount", Invoice_Amount);
        json.put("Invoice_Due_Date", Invoice_Due_Date);
        json.put("Pay_Type", Pay_Type);
        json.put("Pay_Delay", Pay_Delay);
        json.put("Risk", Risk);
        json.put("Model_Name", Model_Name);

        return json ;
    }

    @Override
    public String toString() {
        return "[Invoice_Amount=" + Invoice_Amount + ", Invoice_Due_Date=" + Invoice_Due_Date + ", Invoice_ID="
                + Invoice_ID + ", Pay_Delay=" + Pay_Delay + ", Pay_Type=" + Pay_Type + ", Risk=" + Risk + "]";
    }

}
