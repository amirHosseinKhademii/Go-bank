variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

variable "app_name" {
  description = "Application name"
  type        = string
  default     = "bank-app"
}

variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
  default     = "10.0.0.0/16"
}

variable "container_port" {
  description = "Container port"
  type        = number
  default     = 8080
}

variable "container_image" {
  description = "Container image URI"
  type        = string
}

variable "db_username" {
  description = "Database username"
  type        = string
  sensitive   = true
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

variable "eks_desired_size" {
  description = "Desired size of EKS node group"
  type        = number
  default     = 3
}

variable "eks_min_size" {
  description = "Minimum size of EKS node group"
  type        = number
  default     = 1
}

variable "eks_max_size" {
  description = "Maximum size of EKS node group"
  type        = number
  default     = 10
}

variable "instance_types" {
  description = "Instance types for EKS node group"
  type        = list(string)
  default     = ["t3.medium"]
}

variable "enable_logging" {
  description = "Enable CloudWatch logging"
  type        = bool
  default     = true
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default = {
    Environment = "production"
    ManagedBy   = "terraform"
    Project     = "bank-api"
  }
}
