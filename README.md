<div align="center">
  <img src="https://images.squarespace-cdn.com/content/v1/607861b10c0e3b4816f56581/3eea3ca7-58ca-402d-9edf-b8e9e72eca3c/lightning.png?format=300w" alt="Photon Commerce">

  <h1>BANK & CARD STATEMENTS</h1>
  <p><strong>Extract structured data from bank & card statements with 99.9%+ accuracy</strong></p>
  <p><strong>Powered by <a href="https://www.photoncommerce.com">Photon Commerce</a> — Managed AI Agents with data verification and validation</strong></p>

  <p>
    <a href="https://www.photoncommerce.com"><img src="https://img.shields.io/badge/SOC%202-Compliant-blue?style=for-the-badge" alt="SOC 2"></a>
    &nbsp;&nbsp;
    <a href="https://www.photoncommerce.com"><img src="https://img.shields.io/badge/GDPR-Attested-blue?style=for-the-badge" alt="GDPR"></a>
    &nbsp;&nbsp;
    <a href="https://www.photoncommerce.com/pricing"><img src="https://img.shields.io/badge/Accuracy-99.9%25%2B-00C853?style=for-the-badge" alt="Accuracy"></a>
    &nbsp;&nbsp;
    <a href="https://www.photoncommerce.com/platform"><img src="https://img.shields.io/badge/Languages-25%2B-FF6D00?style=for-the-badge" alt="Languages"></a>
  </p>

  [Website](https://www.photoncommerce.com) &nbsp;·&nbsp; [API Docs](https://apidocs.photoncommerce.com) &nbsp;·&nbsp; [Pricing](https://www.photoncommerce.com/pricing) &nbsp;·&nbsp; [Free Trial](https://app.photoncommerce.com) &nbsp;·&nbsp;[Try in Claude](https://claude.ai/directory/connectors/photoncommerce) &nbsp;·&nbsp; [Try in ChatGPT](https://chatgpt.com/plugins/plugin_asdk_app_696685b735588191b9f25f976cfda7b2)
</div>

---

## Overview

This repository contains ready-to-run code samples for extracting structured data from bank and card statements using the **PhotonCommerce API**.

Submit any statement — checking, savings, or credit card — and receive a structured JSON response with account details, balances, and a full transaction ledger, verified to 99.9%+ accuracy by Photon's Managed Agents.

---

## Extracted Fields

| Category | Fields |
|----------|--------|
| **Account** | Account number, account type, client name, client address |
| **Bank** | Bank name, bank address |
| **Balances** | Starting balance, ending balance, total credits, total debits |
| **Statement Period** | Start date, end date |
| **Transactions** | Date, type (credit/debit), description, amount |
| **Metadata** | Duplicate flag |

Full transaction ledger · Checking, savings & credit card statements · 25+ languages

---

## Quick Start

```python
import requests

HEADERS = {
    "CLIENT-ID":     "YOUR_CLIENT_ID",
    "AUTHORIZATION": "apikey YOUR_USERNAME:YOUR_API_KEY",
    "PASSWORD":      "YOUR_PASSWORD",
    "SECRET-KEY":    "YOUR_SECRET_KEY",
}

# Step 1 — Submit
response = requests.post(
    "https://sandbox-api.photoncommerce.com/api/pro",
    headers=HEADERS,
    params={"doctype": "statement"},
    files={"pdf": open("statement.pdf", "rb")},
)
photon_key = response.json()["photon_key"]

# Step 2 — Retrieve
result = requests.get(
    "https://sandbox-api.photoncommerce.com/api/v4/json",
    headers=HEADERS,
    params={"photon_key": photon_key},
).json()["data"]

print(result["client_name"])        # JANE DOE
print(result["bank_name"])          # CHASE
print(result["account_number"])     # 001100322730721
print(result["starting_balance"])   # 11146.68
print(result["ending_balance"])     # 13946.68
```

[Get your free sandbox credentials →](https://sandbox-api.photoncommerce.com/api/v4/register)

---

## Code Examples

| Language | File | Library |
|----------|------|---------|
| ![Python](https://img.shields.io/badge/Python-3776AB?style=flat-square&logo=python&logoColor=white) | [extract_statement.py](extract_statement.py) | requests |
| ![JavaScript](https://img.shields.io/badge/Node.js-339933?style=flat-square&logo=nodedotjs&logoColor=white) | [extract_statement.js](extract_statement.js) | fetch + form-data |


Every example supports both **local file upload** and **URL-based submission**, plus optional webhook callbacks.

---

## How It Works

```
Bank Statement In (PDF / image)
    ↓
Photon Engine  —  AI + OCR + NLP
    ↓
Managed Agents  —  Validate & Verify
    ↓
Structured JSON Out  —  Account details + full transaction ledger
```

### Submission

```
POST /api/pro?doctype=statement
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `doctype` | string | `statement` |
| `url` | string | URL of a publicly accessible document (alternative to file upload) |
| `webhook_url` | string | Receive a callback when extraction is complete |
| `auth_token` | string | Token to verify the webhook callback |
| `ID` | string | Your own reference ID for this submission |
| `subaccount` | string | Route to a subaccount |
| `page_start` | integer | First page to process (multi-page PDFs) |
| `page_end` | integer | Last page to process (multi-page PDFs) |

### Retrieval

```
GET /api/v4/json?photon_key=YOUR_PHOTON_KEY
```

Or pass `webhook_url` at submission and Photon will POST the result to your endpoint when ready.

---

## Processing Times

| Account Type | Turnaround |
|-------------|------------|
| Trial | Up to 24 hours (Managed Agents included) |
| Production | seconds to 24 hours (Managed Agents included) |

> **Instant Managed Agent:** To activate, contact [support@photoncommerce.com](mailto:support@photoncommerce.com). Once activated, submit to `/api/v4` instead of `/api/pro` for near-instant results.

---

## Authentication

All requests require four headers:

```
CLIENT-ID:     your-client-id
AUTHORIZATION: apikey your-username:your-api-key
PASSWORD:      your-password
SECRET-KEY:    your-secret-key
```

Get credentials: [Register a free sandbox account](https://sandbox-api.photoncommerce.com/api/v4/register) (20 free calls, no credit card required) or sign up at [app.photoncommerce.com](https://app.photoncommerce.com).

---

## Sample Response

```json
{
  "data": {
    "starting_balance": 11146.68,
    "ending_balance": 13946.68,
    "tot_credit": 2802.4,
    "tot_debit": 2.4,
    "statement_end_date": "2021-04-26",
    "statement_start_date": "2021-03-25",
    "account_number": "001100322730721",
    "client_name": "JANE DOE",
    "account_type": "Chase College Checking",
    "bank_name": "CHASE",
    "bank_address": "JPMorgan Chase Bank, N.A.\nPO Box 182051\nColumbus, OH 43218-2051",
    "client_address": "30 NEWPORT PKWY",
    "transactions": [
      {
        "date": "2021-04-01",
        "type": "credit",
        "description": "PPD ID: 9320517957",
        "amount": 2800.0
      },
      {
        "date": "2021-04-16",
        "type": "credit",
        "description": "CCD ID: 1800948598",
        "amount": 1.25
      },
      {
        "date": "2021-04-19",
        "type": "credit",
        "description": "CCD ID: 1800948598",
        "amount": 1.15
      },
      {
        "date": "2021-04-22",
        "type": "debit",
        "description": "American Express ACH Pmt M2730 Web ID: 2005134111",
        "amount": 2.40
      }
    ],
    "photon_key": "data/janedoe/2022-09-26/06-25-44-948850_Chase_Bank_Statement1.json",
    "doc_path": "data/janedoe/2022-09-26/06-25-44-948850_Chase_Bank_Statement1.pdf"
    "Is_Duplicate": 0
  },
  "message": "success",
  "status": "success"
}
```

---

## Other Document Types

| Repo | Document |
|------|----------|
| [`photon-invoice`](https://github.com/Photon-Commerce/photon-invoice) | Invoices |
| [`photon-receipt`](https://github.com/Photon-Commerce/photon-receipt) | Receipts |


---

## Links

| | |
|-|-|
| **Try in Claude** | [Photon Commerce on Claude](https://claude.ai/directory/connectors/photoncommerce) |
| **Try in ChatGPT** | [Photon Commerce on ChatGPT](https://chatgpt.com/plugins/plugin_asdk_app_696685b735588191b9f25f976cfda7b2) |
| **API Docs** | [apidocs.photoncommerce.com](https://apidocs.photoncommerce.com) |
| **Free Trial** | [app.photoncommerce.com](https://app.photoncommerce.com) |
| **Pricing** | [photoncommerce.com/pricing](https://www.photoncommerce.com/pricing) |
| **Support** | [support@photoncommerce.com](mailto:support@photoncommerce.com) |
| **Enterprise** | [developers@photoncommerce.com](mailto:developers@photoncommerce.com) |
