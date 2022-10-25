
resource "aws_iam_role" "lambdamiddleware-examples" {
  name = "lambdamiddleware-examples"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_policy" "lambdamiddleware-examples" {
  name   = "lambdamiddleware-examples"
  path   = "/"
  policy = data.aws_iam_policy_document.lambdamiddleware-examples.json
}

resource "aws_iam_role_policy_attachment" "lambdamiddleware-examples" {
  role       = aws_iam_role.lambdamiddleware-examples.name
  policy_arn = aws_iam_policy.lambdamiddleware-examples.arn
}

data "aws_iam_policy_document" "lambdamiddleware-examples" {
  statement {
    actions = [
      "ssm:GetParameter*",
      "ssm:DescribeParameters",
      "ssm:List*",
    ]
    resources = ["*"]
  }
  statement {
    actions = [
      "logs:GetLog*",
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = ["*"]
  }
}

resource "aws_ssm_parameter" "foo" {
  name        = "/lambdamiddleware-examples/foo"
  description = "foo for lambdamiddleware-examples"
  type        = "String"
  value       = "foo values"
}

resource "aws_ssm_parameter" "bar" {
  name        = "/lambdamiddleware-examples/bar"
  description = "bar for lambdamiddleware-examples"
  type        = "String"
  value       = "bar values"
}
