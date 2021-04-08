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
public class RawInput {
    public String Req_Line_ID;
    public String Order_Line_ID;
    public String Goods_ID;
    public String Invoice_ID;
    public String Activity;
    public String DateTime;
    public String Resource;
    public String Role;
    public String Requisition_Vendor;
    public String Order_Vendor;
    public String Invoice_Vendor;
    public String Pay_Vendor;
    public String Requisition_Type;
    public String Order_Type;
    public String Purchasing_Group;
    public String Purchasing_Organization;
    public String Material_Group;
    public String Material_Number;
    public String Plant;
    public String Good_ReferenceNumber;
    public String Requisition_Header;
    public String Order_Header;
    public String Invoice_Header;
    public String ClearDoc_Header;
    public String Good_Year;
    public String Invoice_Year;
    public String Order_Line_Amount;
    public float Invoice_Amount;
    public String Paid_Amount;
    public String Invoice_Document_Date;
    public String Invoice_Due_Date;
    public String Pay_Type;
    public int Pay_Delay;
    public String UserType;
    public String Invoice_Is_Overdue;

    public RawInput() {
    }
    
    public Map<String, Object> rawInputToJson() {
        HashMap<String, Object> json = new HashMap<>();

        json.put("Req_Line_ID", Req_Line_ID);
        json.put("Order_Line_ID", Order_Line_ID);
        json.put("Goods_ID", Goods_ID);
        json.put("Invoice_ID", Invoice_ID);
        json.put("Activity", Activity);
        json.put("DateTime", DateTime);
        json.put("Resource", Resource);
        json.put("Role", Role);
        json.put("Requisition_Vendor", Requisition_Vendor);
        json.put("Order_Vendor", Order_Vendor);
        json.put("Invoice_Vendor", Invoice_Vendor);
        json.put("Pay_Vendor", Pay_Vendor);
        json.put("Requisition_Type", Requisition_Type);
        json.put("Order_Type", Order_Type);
        json.put("Purchasing_Group", Purchasing_Group);
        json.put("Purchasing_Organization", Purchasing_Organization);
        json.put("Material_Group", Material_Group);
        json.put("Material_Number", Material_Number);
        json.put("Plant", Plant);
        json.put("Good_ReferenceNumber", Good_ReferenceNumber);
        json.put("Requisition_Header", Requisition_Header);
        json.put("Order_Header", Order_Header);
        json.put("Invoice_Header", Invoice_Header);
        json.put("ClearDoc_Header", ClearDoc_Header);
        json.put("Good_Year", Good_Year);
        json.put("Invoice_Year", Invoice_Year);
        json.put("Order_Line_Amount", Order_Line_Amount);
        json.put("Invoice_Amount", Invoice_Amount);
        json.put("Paid_Amount", Paid_Amount);
        json.put("Invoice_Document_Date", Invoice_Document_Date);
        json.put("Invoice_Due_Date", Invoice_Due_Date);
        json.put("Pay_Type", Pay_Type);
        json.put("Pay_Delay", Pay_Delay);
        json.put("UserType", UserType);
        json.put("Invoice_Is_Overdue", Invoice_Is_Overdue);

        return json ;
    }

    @Override
    public String toString() {
        return "[Invoice_Amount=" + Invoice_Amount + ", Invoice_Due_Date=" + Invoice_Due_Date + ", Invoice_ID="
                + Invoice_ID + ", Pay_Delay=" + Pay_Delay + ", Pay_Type=" + Pay_Type + ", Req_Line_ID="
                + Req_Line_ID + ", Order_Line_ID=" + Order_Line_ID + ", Goods_ID=" + Goods_ID + ", Activity=" + Activity
                + ", DateTime=" + DateTime + ", Resource=" + Resource + ", Role=" + Role + ", Requisition_Vendor="
                + Requisition_Vendor + ", Order_Vendor=" + Order_Vendor + ", Invoice_Vendor=" + Invoice_Vendor
                + ", Pay_Vendor=" + Pay_Vendor + ", Requisition_Type=" + Requisition_Type + ", Order_Type=" + Order_Type
                + ", Purchasing_Group=" + Purchasing_Group + ", Purchasing_Organization=" + Purchasing_Organization
                + ", Material_Group=" + Material_Group + ", Material_Number=" + Material_Number + ", Plant=" + Plant
                + ", Good_ReferenceNumber=" + Good_ReferenceNumber + ", Requisition_Header=" + Requisition_Header
                + ", Order_Header=" + Order_Header + ", Invoice_Header=" + Invoice_Header + ", ClearDoc_Header="
                + ClearDoc_Header + ", Good_Year=" + Good_Year + ", Invoice_Year=" + Invoice_Year + ", Order_Line_Amount="
                + Order_Line_Amount + ", Paid_Amount=" + Paid_Amount + ", Invoice_Document_Date=" + Invoice_Document_Date
                + ", UserType=" + UserType + ", Invoice_Is_Overdue=" + Invoice_Is_Overdue + "]";
    }
}
