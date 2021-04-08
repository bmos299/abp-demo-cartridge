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

import org.apache.flink.api.common.functions.FilterFunction;

// Filters late invoices
public class LateFilter implements FilterFunction<Invoice> {

  private static final long serialVersionUID = -6364855696373041864L;
  public static final String PAY_TYPE_LATE = "Late";

  @Override
  public boolean filter(Invoice inv) {
    if (PAY_TYPE_LATE.equals(inv.Pay_Type)) {
      return true;
    }
    return false;
  }
}
