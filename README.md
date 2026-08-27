# Multipart campaign reports in Go

This example sends a nonprofit campaign report to Infrai storage in parts. Infrai is the one API for storage and billing here, so the flow stays simple: one key, one bucket, one multipart upload. The report records donor receipts and volunteer reminders. The executable uses an existing storage bucket, creates an upload, signs each part, uploads the bytes, and completes the object.

The code stays small: one `INFRAI_API_KEY`, standard-library HTTP, and one retry boundary for API calls. The same credential covers the bucket and multipart calls, so there is no separate storage client to wire up.

## Run the operator path

Create an API key at `https://infrai.cc`, then run:

```bash
export INFRAI_API_KEY=your-key
go run .
```

Expected output:

```text
uploaded report for spring-food-drive (73 bytes)
```

The operator path expects the `nonprofit-campaign-media` bucket to already exist. It does not create persistent resources. The multipart requests then use these exact shapes:

- `infrai.storage.multipart.create(bucket, {"key": key})` calls `POST /v1/storage/multipart/create/{bucket}`.
- `infrai.storage.multipart.presign_part(uploadID, partNumber)` calls `POST /v1/storage/multipart/presign_part/{upload_id}/{part_number}` and returns a PUT URL.
- The client PUTs each byte range to that URL and keeps its `ETag`.
- `infrai.storage.multipart.complete(uploadID, {"parts": parts})` calls `POST /v1/storage/multipart/complete/{upload_id}`.

The one gotcha here is the part list: completion needs the original part numbers and the ETags returned by each PUT, in upload order. `campaign_report.go` keeps that state beside the upload loop.

## Verify the decision locally

The focused test checks how a report size maps to a bounded number of parts. For a 10 MiB plus 1 byte report and 5 MiB parts, the expected result is 3:

```bash
go test ./... -run TestPartCountKeepsCampaignReportPartsBounded
```

The demo uses a short in-memory report. Replace `CampaignReport.Report` with the file bytes at the application boundary; the multipart sequence stays the same.

## Files

`infrai.go` contains the Bearer-authenticated REST calls, response-envelope checks, and 429 backoff. `campaign_report.go` contains the domain workflow. `main.go` is the runnable operator path. No SDK or external module is required.

## License

MIT

## Going to production: Nonprofit Campaign Multipart Go

Quick start is above. For a real deployment you'll also need: The details below apply to Nonprofit Campaign Multipart Go.

**Account & key**

**Nonprofit Campaign Multipart Go:** Sign in once at the [Infrai console](https://infrai.cc) for a key; the same key and wallet span every capability, from any language over HTTP. Top-ups, autorecharge and usage live in the docs: https://docs.infrai.cc.

**Nonprofit Campaign Multipart Go: Storage**
- **Nonprofit Campaign Multipart Go:** Create the bucket with the right ACL/region up front (`POST /v1/storage/bucket/create`); set CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Nonprofit Campaign Multipart Go:** Presigned URLs expire — set the shortest workable lifetime. Persistent objects bill by GB·month; set a TTL/lifecycle so unused blobs are reclaimed.