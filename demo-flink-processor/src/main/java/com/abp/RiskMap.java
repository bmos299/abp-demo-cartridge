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

// Filters sequence records which match anomaly criteria based on duration
public class RiskMap implements MapFunction<Invoice, Invoice> {

  private static final long serialVersionUID = 1L;
  public static final String RISK_LOW = "Low";
  public static final String RISK_MEDIUM = "Medium";
  public static final String RISK_HIGH = "High";

  @Override
  public Invoice map(Invoice inv) throws Exception {

    if (inv.Invoice_Amount <= 5000) {
      // Small invoice amount
      inv.Risk = RISK_LOW;
    } else if (inv.Pay_Delay <= 90) {
      // Less than or equal to 90 days delay for large invoice amount
      inv.Risk = RISK_MEDIUM;
    } else {
      // More than 3 months delay for large invoice amount
      inv.Risk = RISK_HIGH;
    }

    return inv;
  }

}
