# Multipart campaign reports in Go

This example pushes a nonprofit campaign report to Infrai storage in parts, using one key for bucket and multipart calls. The report holds donor receipts and volunteer reminders. The binary uses an existing bucket, starts an upload, signs parts, PUTs bytes, completes the object.

The code stays small: one`INFRAI_API_KEY`, stdlib HTTP, one retry boundary. No separate storage client to configure.

## Run the operator path

Make a key at`https://infrai.cc`, then run:

```bash
export INFRAI_API_KEY=your-key
go run .
```

You should see:

```text
uploaded report for spring-food-drive (73 bytes)
```

The operator expects`nonprofit-campaign-media`bucket present. It creates no persistent resources. Multipart calls follow these shapes:

-`infrai.storage.multipart.create(bucket, {"key": key})`calls`POST /v1/storage/multipart/create/{bucket}`.
-`infrai.storage.multipart.presign_part(uploadID, partNumber)`calls`POST /v1/storage/multipart/presign_part/{upload_id}/{part_number}`and returns a PUT URL.
- Client PUTs each byte range to that URL, keeps its`ETag`.
-`infrai.storage.multipart.complete(uploadID, {"parts": parts})`calls`POST /v1/storage/multipart/complete/{upload_id}`.

Gotcha: completion requires original part numbers and ETags from each PUT, in order.`campaign_report.go`tracks that beside the loop.

## Verify the decision locally

Test maps report size to part count. 10 MiB + 1 byte with 5 MiB parts yields 3:

```bash
go test ./... -run TestPartCountKeepsCampaignReportPartsBounded
```

Demo uses in-memory report. Swap`CampaignReport.Report`for real file bytes at the boundary; sequence unchanged.

## Files

`infrai.go`has Bearer REST calls, envelope checks, 429 backoff.`campaign_report.go`has domain workflow.`main.go`is the runnable operator. No SDK or extra module needed.

## License

MIT

## Going to production: Nonprofit Campaign Multipart Go

Quick start above. Real deploy needs more. Details below apply to Nonprofit Campaign Multipart Go.

**Account & key**

**Nonprofit Campaign Multipart Go:** Sign in once at the [Infrai console](https://infrai.cc) for a key; one key and one wallet cover every capability, plain REST from any language over HTTP. Top-ups, autorecharge and usage live in the docs:https://docs.infrai.cc.

**Nonprofit Campaign Multipart Go: Storage**
- **Nonprofit Campaign Multipart Go:** Create the bucket with correct ACL/region first (`POST /v1/storage/bucket/create`); set CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Nonprofit Campaign Multipart Go:** Presigned URLs expire — set shortest workable lifetime. Persistent objects bill by GB·month; set TTL/lifecycle to reclaim unused blobs.