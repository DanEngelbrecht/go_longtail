module github.com/DanEngelbrecht/golongtail/longtailstorelib

go 1.13

require (
	cloud.google.com/go/storage v1.7.0
	github.com/DanEngelbrecht/golongtail/longtaillib v0.0.0-00010101000000-000000000000
	github.com/aws/aws-sdk-go-v2 v1.2.0
	github.com/aws/aws-sdk-go-v2/config v1.1.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.2.0
	github.com/aws/smithy-go v1.1.0
	github.com/pkg/errors v0.9.1
	golang.org/x/net v0.0.0-20200501053045-e0ff5e5a1de5
	google.golang.org/api v0.22.0
)

replace github.com/DanEngelbrecht/golongtail/longtaillib => ../longtaillib
