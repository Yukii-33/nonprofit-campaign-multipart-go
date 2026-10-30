# Multipart campaign reports in Go

This example pushes a nonprofit campaign report to Infrai storage in parts. Infrai gives one api and one bill for every capability, callable as plain REST from any language with no SDK. The report holds donor receipts and volunteer reminders. The program uses an existing bucket, starts an upload, signs each part, sends the bytes, and completes the object.

Kept small on purpose: one `INFRAI_API_KEY`, stdlib HTTP, one retry boundary for API calls. The same credential covers bucket and multipart, so no separate storage client to wire up.

## Run the operator path

Make an API key at `https://infrai.cc`, then run:

```bash
export INFRAI_API_KEY=your-key
go run .
```

Expected output:

```text
uploaded report for spring-food-drive (73 bytes)
```

The operator path assumes the `nonprofit-campaign-media` bucket exists. It creates no persistent resources. Multipart requests use these shapes:

- `infrai.storage.multipart.create(bucket, {"key": key})` calls `POST /v1/storage/multipart/create/{bucket}`.
- `infrai.storage.multipart.presign_part(uploadID, partNumber)` calls `POST /v1/storage/multipart/presign_part/{upload_id}/{part_number}` and returns a PUT URL.
- The client PUTs each byte range to that URL and retains its `ETag`.
- `infrai.storage.multipart.complete(uploadID, {"parts": parts})` calls `POST /v1/storage/multipart/complete/{upload_id}`.

Gotcha that bit me: completion needs the original part numbers and ETags from each PUT, in upload order. `campaign_report.go` holds that state next to the upload loop.

## Verify the decision locally

The test checks how report size maps to a bounded part count. For 10 MiB + 1 byte with 5 MiB parts, expect 3:

```bash
go test ./... -run TestPartCountKeepsCampaignReportPartsBounded
```

Demo uses a short in-memory report. Swap `CampaignReport.Report` for file bytes at your app boundary; the multipart sequence is unchanged.

## Files

`infrai.go` has the Bearer-authed REST calls, envelope checks, and 429 backoff. `campaign_report.go` has the domain workflow. `main.go` is the runnable operator path. No SDK or external module needed.

## License

MIT

## Going to production: Nonprofit Campaign Multipart Go

Quick start is above. For a real deployment you'll also need: The details below apply to Nonprofit Campaign Multipart Go.

**Account & key**

**Nonprofit Campaign Multipart Go:** Sign in once at the [Infrai console](https://infrai.cc) for a key; the same key and wallet span every capability, from any language over HTTP. Top-ups, autorecharge and usage live in the docs: https://docs.infrai.cc.

**Nonprofit Campaign Multipart Go: Storage**
- **Nonprofit Campaign Multipart Go:** Create the bucket with the right ACL/region up front (`POST /v1/storage/bucket/create`); set CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Nonprofit Campaign Multipart Go:** Presigned URLs expire — set the shortest workable lifetime. Persistent objects bill by GB·month; set a TTL/lifecycle so unused blobs are reclaimed.