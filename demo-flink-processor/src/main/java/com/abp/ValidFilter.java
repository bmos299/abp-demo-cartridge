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

// Filters valid invoices
public class ValidFilter implements FilterFunction<Invoice> {

  private static final long serialVersionUID = -6364855696373041864L;

  @Override
  public boolean filter(Invoice inv) {
    if (inv.Invoice_Amount > 0) {
      return true;
    }
    return false;
  }
}
