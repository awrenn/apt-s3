provider "aws" {
  region = "us-west-2"
}


resource "aws_s3_bucket" "dummy_apt_repo" {
  bucket = "dummy-apt"
  acl    = "public-read"
}

resource "aws_s3_bucket_policy" "allow_access_from_another_account" {
  bucket = aws_s3_bucket.dummy_apt_repo.id
  policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
    {
        "Sid": "PublicReadGetObject",
        "Effect": "Allow",
        "Principal": "*",
        "Action": "s3:GetObject",
        "Resource": "${aws_s3_bucket.dummy_apt_repo.arn}/*"
    }
 ]
}
EOF
}

output "dummy_repo_url" {
    value = aws_s3_bucket.dummy_apt_repo.bucket_regional_domain_name
}
