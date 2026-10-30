/**
 * Extract structured data from bank and card statements using the Photon Commerce API.
 *
 * Submits a statement (PDF or image) and returns account details, balances,
 * and a full transaction ledger including client name, bank name, account number,
 * starting/ending balances, and all transactions (date, type, description, amount).
 *
 * Processing times (Managed Agents):
 *   Trial accounts:  up to 24 hours
 *   Production:      5 minutes to 24 hours
 *
 * AI extraction (seconds, no Managed Agents):
 *   Contact support@photoncommerce.com to activate.
 *   Once active, submit to /api/v4 instead of /api/pro.
 *
 * Requires: OkHttp 4.x, org.json
 *
 * Docs:    https://apidocs.photoncommerce.com
 * Sandbox: https://sandbox-api.photoncommerce.com/api/v4/register (20 free calls)
 */

import okhttp3.*;
import org.json.JSONArray;
import org.json.JSONObject;

import java.io.File;
import java.io.IOException;
import java.util.concurrent.TimeUnit;

public class ExtractBankStatement {

    private static final String CLIENT_ID  = "YOUR_CLIENT_ID";
    private static final String USERNAME   = "YOUR_USERNAME";
    private static final String API_KEY    = "YOUR_API_KEY";
    private static final String PASSWORD   = "YOUR_PASSWORD";
    private static final String SECRET_KEY = "YOUR_SECRET_KEY";

    // Sandbox: https://sandbox-api.photoncommerce.com  (20 free calls, no card needed)
    // Production: https://api.photoncommerce.com
    private static final String BASE_URL = "https://sandbox-api.photoncommerce.com";

    private static final OkHttpClient client = new OkHttpClient.Builder()
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .build();

    private static Request.Builder authBuilder() {
        return new Request.Builder()
                .header("CLIENT-ID", CLIENT_ID)
                .header("AUTHORIZATION", "apikey " + USERNAME + ":" + API_KEY)
                .header("PASSWORD", PASSWORD)
                .header("SECRET-KEY", SECRET_KEY);
    }

    public static String submitStatement(
            String filePath, String url,
            String webhookUrl, String authToken,
            String id, String subaccount,
            Integer pageStart, Integer pageEnd) throws IOException {

        if (filePath == null && url == null)
            throw new IllegalArgumentException("Provide either filePath or url.");

        // For AI extraction (seconds), replace /api/pro with /api/v4 — contact support@photoncommerce.com to activate.
        HttpUrl.Builder urlBuilder = HttpUrl.parse(BASE_URL + "/api/pro").newBuilder()
                .addQueryParameter("doctype", "statement");
        if (url != null)        urlBuilder.addQueryParameter("url", url);
        if (webhookUrl != null) urlBuilder.addQueryParameter("webhook_url", webhookUrl);
        if (authToken != null)  urlBuilder.addQueryParameter("auth_token", authToken);
        if (id != null)         urlBuilder.addQueryParameter("ID", id);
        if (subaccount != null) urlBuilder.addQueryParameter("subaccount", subaccount);
        if (pageStart != null)  urlBuilder.addQueryParameter("page_start", String.valueOf(pageStart));
        if (pageEnd != null)    urlBuilder.addQueryParameter("page_end", String.valueOf(pageEnd));

        RequestBody body;
        if (filePath != null) {
            body = new MultipartBody.Builder().setType(MultipartBody.FORM)
                    .addFormDataPart("pdf", new File(filePath).getName(),
                            RequestBody.create(new File(filePath), MediaType.parse("application/octet-stream")))
                    .build();
        } else {
            body = RequestBody.create(new byte[0]);
        }

        Request request = authBuilder().url(urlBuilder.build()).post(body).build();
        try (Response response = client.newCall(request).execute()) {
            return new JSONObject(response.body().string()).getString("photon_key");
        }
    }

    public static JSONObject fetchResult(String photonKey) throws IOException {
        HttpUrl url = HttpUrl.parse(BASE_URL + "/api/v4/json").newBuilder()
                .addQueryParameter("photon_key", photonKey).build();
        Request request = authBuilder().url(url).get().build();
        try (Response response = client.newCall(request).execute()) {
            return new JSONObject(response.body().string()).optJSONObject("data");
        }
    }

    public static JSONObject waitForResult(String photonKey, int pollIntervalSec, int timeoutSec)
            throws IOException, InterruptedException {
        long deadline = System.currentTimeMillis() + (long) timeoutSec * 1000;
        while (System.currentTimeMillis() < deadline) {
            JSONObject result = fetchResult(photonKey);
            String status = result != null ? result.optString("Status", "") : "";
            if (!status.isEmpty() && !status.equals("pending") && !status.equals("processing")) return result;
            System.out.println("  Status: " + (status.isEmpty() ? "pending" : status)
                    + " — retrying in " + pollIntervalSec + "s...");
            Thread.sleep(pollIntervalSec * 1000L);
        }
        throw new RuntimeException("Extraction not complete after " + timeoutSec + "s");
    }

    public static void main(String[] args) throws Exception {
        // --- Option A: submit from a local file ---
        String photonKey = submitStatement("statement.pdf", null, null, null, null, null, null, null);

        // --- Option B: submit via a publicly accessible URL ---
        // String photonKey = submitStatement(null, "https://example.com/statement.pdf", null, null, null, null, null, null);

        System.out.println("Submitted. photon_key: " + photonKey);
        System.out.println("Waiting for extraction to complete...");

        JSONObject result = waitForResult(photonKey, 20, 3600);

        System.out.println("\n--- Bank Statement Data ---");
        System.out.println("Client:           " + result.optString("client_name"));
        System.out.println("Bank:             " + result.optString("bank_name"));
        System.out.println("Account Number:   " + result.optString("account_number"));
        System.out.println("Account Type:     " + result.optString("account_type"));
        System.out.println("Period:           " + result.optString("statement_start_date") + " → " + result.optString("statement_end_date"));
        System.out.println("Starting Balance: " + result.opt("starting_balance"));
        System.out.println("Ending Balance:   " + result.opt("ending_balance"));
        System.out.println("Total Credits:    " + result.opt("tot_credit"));
        System.out.println("Total Debits:     " + result.opt("tot_debit"));

        JSONArray transactions = result.optJSONArray("transactions");
        if (transactions != null && transactions.length() > 0) {
            System.out.println("\n--- Transactions (" + transactions.length() + ") ---");
            for (int i = 0; i < transactions.length(); i++) {
                JSONObject t = transactions.getJSONObject(i);
                System.out.printf("  %-12s %-6s %10s  %s%n",
                        t.optString("date"), t.optString("type"),
                        t.opt("amount"), t.optString("description"));
            }
        }
    }
}
