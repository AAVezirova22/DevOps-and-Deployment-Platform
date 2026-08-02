variable "hcloud_token" {
  description = "Hetzner Cloud API token."
  type        = string
  sensitive   = true
}

variable "project_name" {
  description = "Name prefix for created infrastructure."
  type        = string
  default     = "deploykit"
}

variable "ssh_public_key" {
  description = "Public SSH key allowed to access the server."
  type        = string
}

variable "admin_cidrs" {
  description = "CIDR ranges allowed to SSH and access the Kubernetes API."
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "server_location" {
  description = "Hetzner location, such as fsn1, nbg1, hel1, ash, hil, or sin."
  type        = string
  default     = "fsn1"
}

variable "server_type" {
  description = "Hetzner server type."
  type        = string
  default     = "cx22"
}

variable "k3s_channel" {
  description = "k3s release channel."
  type        = string
  default     = "stable"
}
