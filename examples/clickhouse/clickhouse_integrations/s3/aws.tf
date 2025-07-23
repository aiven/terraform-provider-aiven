# S3 bucket for ClickHouse data
resource "aws_s3_bucket" "clickhouse_data" {
  bucket = "${var.service_name_prefix}-clickhouse-data"

  tags = {
    Name        = "ClickHouse Data Bucket"
    Environment = "example"
  }
}

# S3 bucket server-side encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "clickhouse_data" {
  bucket = aws_s3_bucket.clickhouse_data.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Block public access to S3 bucket
resource "aws_s3_bucket_public_access_block" "clickhouse_data" {
  bucket = aws_s3_bucket.clickhouse_data.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# IAM policy for ClickHouse to access S3
resource "aws_iam_policy" "clickhouse_s3_access" {
  name        = "${var.service_name_prefix}-clickhouse-s3-access"
  description = "Policy for ClickHouse to access S3 bucket"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket",
          "s3:GetBucketLocation"
        ]
        Resource = [
          aws_s3_bucket.clickhouse_data.arn,
          "${aws_s3_bucket.clickhouse_data.arn}/*"
        ]
      }
    ]
  })
}

# IAM user for ClickHouse
resource "aws_iam_user" "clickhouse_user" {
  name = "${var.service_name_prefix}-clickhouse-user"
  path = "/"

  tags = {
    Environment = "example"
    Service     = "clickhouse"
  }
}

# IAM access key for the ClickHouse user
resource "aws_iam_access_key" "clickhouse_user" {
  user = aws_iam_user.clickhouse_user.name
}

# Attach policy to user
resource "aws_iam_user_policy_attachment" "clickhouse_s3_access" {
  user       = aws_iam_user.clickhouse_user.name
  policy_arn = aws_iam_policy.clickhouse_s3_access.arn
}

# Sample CSV data for testing the S3 integration
resource "aws_s3_object" "test_csv" {
  bucket = aws_s3_bucket.clickhouse_data.bucket
  key    = "test_data.csv"

  content = <<-EOT
id,name,email,created_at,status
1,John Doe,john.doe@example.com,2024-01-15 10:30:00,active
2,Jane Smith,jane.smith@example.com,2024-01-16 11:45:00,active
3,Bob Johnson,bob.johnson@example.com,2024-01-17 09:15:00,inactive
4,Alice Brown,alice.brown@example.com,2024-01-18 14:20:00,active
5,Charlie Wilson,charlie.wilson@example.com,2024-01-19 16:30:00,active
EOT

  content_type = "text/csv"

  tags = {
    Purpose = "ClickHouse Test Data"
    Type    = "Sample Users"
  }
}
