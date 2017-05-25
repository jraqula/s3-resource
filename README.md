# S3 Resource

Versions objects in an S3 bucket, by pattern-matching filenames to identify
version numbers.

## Source Configuration

* `bucket`: *Required.* The name of the bucket.

* `signatures_bucket`: *Required.* The signature file for each detected/created
   blob will appear in this bucket.

* `signature_postfix`: *Optional.* The signature file for each detected/created
   blob will have this string appended to the filename of the original blob.

* `access_key_id`: *Optional.* The AWS access key to use when accessing the
  bucket.

* `secret_access_key`: *Optional.* The AWS secret key to use when accessing
  the bucket.

* `region_name`: *Optional.* The region the bucket is in. Defaults to
  `us-east-1`.

* `private`: *Optional.* Indicates that the bucket is private, so that any
  URLs provided are signed.

* `cloudfront_url`: *Optional.* The URL (scheme and domain) of your CloudFront
  distribution that is fronting this bucket (e.g
  `https://d5yxxxxx.cloudfront.net`).  This will affect `in` but not `check`
  and `put`. `in` will ignore the `bucket` name setting, exclusively using the
  `cloudfront_url`.  When configuring CloudFront with versioned buckets, set
  `Query String Forwarding and Caching` to `Forward all, cache based on all` to
  ensure S3 calls succeed.

* `endpoint`: *Optional.* Custom endpoint for using S3 compatible provider.

* `disable_ssl`: *Optional.* Disable SSL for the endpoint, useful for S3
  compatible providers without SSL.

* `server_side_encryption`: *Optional.* An encryption algorithm to use when
  storing objects in S3.

* `sse_kms_key_id`: *Optional.* The ID of the AWS KMS master encryption key
  used for the object.

* `use_v2_signing`: *Optional.* Use signature v2 signing, useful for S3 compatible providers that do not support v4.

### File Names

One of the following two options must be specified:

* `regexp`: *Optional.* The pattern to match filenames against within S3. The first
  grouped match is used to extract the version, or if a group is explicitly
  named `version`, that group is used. At least one capture group must be
  specified, with parentheses.

  The version extracted from this pattern is used to version the resource.
  Semantic versions, or just numbers, are supported. Accordingly, full regular
  expressions are supported, to specify the capture groups.

* `versioned_file`: *Optional* If you enable versioning for your S3 bucket then
  you can keep the file name the same and upload new versions of your file
  without resorting to version numbers. This property is the path to the file
  in your S3 bucket.

## Behavior

### `check`: Extract versions from the bucket.

Objects will be found via the pattern configured by `regexp`. The versions
will be used to order them (using [semver](http://semver.org/)). Each
object's filename is the resulting version.

A found object will only report a version if there is a matching file in
`signatures_bucket` whose filename is the object's name postfixed with
`signature_postfix`.

### `in`: Fetch an object from the bucket.

A blob and signature file will be downloaded, the latter used to verify the
authenticity of the blob's bits using the `public_key` param. If validation
fails, the `get` step will error. The signature file is not persisted on the
rendered volume.

Places the following files in the destination:

* `(filename)`: The file fetched from the bucket.

* `url`: A file containing the URL of the object. If `private` is true, this
  URL will be signed.

* `version`: The version identified in the file name.

#### Parameters

* `public_key`: *Required.* The public key PEM to verify signatures with.

### `out`: Upload an object to the bucket.

Given a file specified by `file`, upload it to the S3 bucket. If `regexp` is
specified, the new file will be uploaded to the directory that the regex
searches in. If `versioned_file` is specified, the new file will be uploaded as
a new version of that file.

Before uploading, the file will be signed and a signature file will be uploaded
to `signatures_bucket`. If your `signatures_bucket` is the same as your `bucket`
you must provide a `signature_postfix` to distinguish the signature's filename.
The signing key provided in the `private_key` param will be used for signing.

#### Parameters

* `file`: *Required.* Path to the file to upload, provided by an output of a task.
  If multiple files are matched by the glob, an error is raised. The file which
  matches will be placed into the directory structure on S3 as defined in `regexp`
  in the resource definition. The matching syntax is bash glob expansion, so
  no capture groups, etc.

* `private_key` *Required.* The private key in PEM format to sign the blob with.

* `acl`: *Optional.*  [Canned Acl](http://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html)
  for the uploaded object.

* `content_type`: *Optional.* MIME [Content-Type](https://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.17)
  describing the contents of the uploaded object

## Example Configuration

### Resource

``` yaml
- name: release
  type: s3
  source:
    bucket: releases
    signatures_bucket: releases_authority
    regexp: directory_on_s3/release-(.*).tgz
    access_key_id: ACCESS-KEY
    secret_access_key: SECRET
```

### Plan

``` yaml
- get: release
  params:
    public_key: |
      -----BEGIN PUBLIC KEY-----
      ...
      -----END PUBLIC KEY-----
```

``` yaml
- put: release
  params:
    file: path/to/release-*.tgz
    public_key: |
      -----BEGIN PRIVATE KEY-----
      ...
      -----END PRIVATE KEY-----
    acl: public-read
```

## Required IAM Permissions

### Non-versioned Buckets

The bucket itself (e.g. `"arn:aws:s3:::your-bucket"`):
* `s3:ListBucket`

The objects in the bucket (e.g. `"arn:aws:s3:::your-bucket/*"`):
* `s3:PutObject`
* `s3:PutObjectAcl`
* `s3:GetObject`

### Versioned Buckets

Everything above and...

The bucket itself (e.g. `"arn:aws:s3:::your-bucket"`):
* `s3:ListBucketVersions`
* `s3:GetBucketVersioning`

The objects in the bucket (e.g. `"arn:aws:s3:::your-bucket/*"`):
* `s3:GetObjectVersion`
* `s3:PutObjectVersionAcl`

## Developing on this resource

First get the resource via:
`go get github.com/concourse/s3-resource`


