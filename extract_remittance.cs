// Extract structured data from remittance advice documents using the Photon Commerce API.
//
// Submits a remittance advice (PDF or image) and returns 30+ structured fields including
// vendor name, customer name, payment date, payment method, reference number, total,
// and a full line-item breakdown for every invoice.
//
// Processing times (Managed Agents):
//   Trial accounts:  up to 24 hours
//   Production:      5 minutes to 24 hours
//
// AI extraction (seconds, no Managed Agents):
//   Contact support@photoncommerce.com to activate.
//   Once active, submit to /api/v4 instead of /api/pro.
//
// Requires: .NET 6+
//
// Docs:    https://apidocs.photoncommerce.com
// Sandbox: https://sandbox-api.photoncommerce.com/api/v4/register (20 free calls)

using System;
using System.Collections.Generic;
using System.IO;
using System.Net.Http;
using System.Net.Http.Json;
using System.Text.Json;
using System.Threading.Tasks;

const string ClientId  = "YOUR_CLIENT_ID";
const string Username  = "YOUR_USERNAME";
const string ApiKey    = "YOUR_API_KEY";
const string Password  = "YOUR_PASSWORD";
const string SecretKey = "YOUR_SECRET_KEY";

// Sandbox: https://sandbox-api.photoncommerce.com  (20 free calls, no card needed)
// Production: https://api.photoncommerce.com
const string BaseUrl = "https://sandbox-api.photoncommerce.com";

using var http = new HttpClient();
http.DefaultRequestHeaders.Add("CLIENT-ID", ClientId);
http.DefaultRequestHeaders.Add("AUTHORIZATION", $"apikey {Username}:{ApiKey}");
http.DefaultRequestHeaders.Add("PASSWORD", Password);
http.DefaultRequestHeaders.Add("SECRET-KEY", SecretKey);

async Task<string> SubmitRemittance(
    string filePath = null, string url = null,
    string webhookUrl = null, string authToken = null,
    string id = null, string subaccount = null,
    int? pageStart = null, int? pageEnd = null)
{
    if (filePath == null && url == null)
        throw new ArgumentException("Provide either filePath or url.");

    var query = new List<string> { "doctype=remittance" };
    if (url != null)        query.Add($"url={Uri.EscapeDataString(url)}");
    if (webhookUrl != null) query.Add($"webhook_url={Uri.EscapeDataString(webhookUrl)}");
    if (authToken != null)  query.Add($"auth_token={Uri.EscapeDataString(authToken)}");
    if (id != null)         query.Add($"ID={Uri.EscapeDataString(id)}");
    if (subaccount != null) query.Add($"subaccount={Uri.EscapeDataString(subaccount)}");
    if (pageStart != null)  query.Add($"page_start={pageStart}");
    if (pageEnd != null)    query.Add($"page_end={pageEnd}");

    // For AI extraction (seconds), replace /api/pro with /api/v4 — contact support@photoncommerce.com to activate.
    var endpoint = $"{BaseUrl}/api/pro?{string.Join("&", query)}";

    HttpContent content;
    if (filePath != null)
    {
        var form = new MultipartFormDataContent();
        var fileBytes = await File.ReadAllBytesAsync(filePath);
        form.Add(new ByteArrayContent(fileBytes), "pdf", Path.GetFileName(filePath));
        content = form;
    }
    else { content = new StringContent(""); }

    var response = await http.PostAsync(endpoint, content);
    response.EnsureSuccessStatusCode();
    var json = await response.Content.ReadFromJsonAsync<JsonElement>();
    return json.GetProperty("photon_key").GetString();
}

async Task<JsonElement> FetchResult(string photonKey)
{
    var response = await http.GetAsync($"{BaseUrl}/api/v4/json?photon_key={photonKey}");
    response.EnsureSuccessStatusCode();
    var json = await response.Content.ReadFromJsonAsync<JsonElement>();
    return json.GetProperty("data");
}

async Task<JsonElement> WaitForResult(string photonKey, int pollIntervalSec = 20, int timeoutSec = 3600)
{
    var deadline = DateTime.UtcNow.AddSeconds(timeoutSec);
    while (DateTime.UtcNow < deadline)
    {
        var result = await FetchResult(photonKey);
        var status = result.TryGetProperty("Status", out var s) ? s.GetString() : null;
        if (!string.IsNullOrEmpty(status) && status != "pending" && status != "processing") return result;
        Console.WriteLine($"  Status: {status ?? "pending"} — retrying in {pollIntervalSec}s...");
        await Task.Delay(pollIntervalSec * 1000);
    }
    throw new TimeoutException($"Extraction not complete after {timeoutSec}s");
}

string str(JsonElement el, string key) =>
    el.TryGetProperty(key, out var v) ? v.GetString() ?? "" : "";

// --- Option A: submit from a local file ---
var photonKey = await SubmitRemittance(filePath: "remittance.pdf");

// --- Option B: submit via a publicly accessible URL ---
// var photonKey = await SubmitRemittance(url: "https://example.com/remittance.pdf");

Console.WriteLine($"Submitted. photon_key: {photonKey}");
Console.WriteLine("Waiting for extraction to complete...");

var result = await WaitForResult(photonKey);

Console.WriteLine("\n--- Remittance Data ---");
Console.WriteLine($"Payee:          {str(result, "payeeName")}");
Console.WriteLine($"Payer:          {str(result, "payerName")}");
Console.WriteLine($"Date:           {str(result, "documentDate")}");
Console.WriteLine($"Payment Method: {str(result, "paymentMethod")}");
Console.WriteLine($"Check Number:   {str(result, "checkNumber")}");
Console.WriteLine($"Total:          {str(result, "total")} {str(result, "currencyCode")}");

if (result.TryGetProperty("lineItems", out var lineItems) && lineItems.GetArrayLength() > 0)
{
    Console.WriteLine($"\n--- Line Items ({lineItems.GetArrayLength()}) ---");
    foreach (var item in lineItems.EnumerateArray())
        Console.WriteLine($"  Invoice {str(item, "invoiceNumber")}: " +
            $"{str(item, "description")}, " +
            $"amount {str(item, "amount")}, " +
            $"discount {str(item, "discount")}");
}
