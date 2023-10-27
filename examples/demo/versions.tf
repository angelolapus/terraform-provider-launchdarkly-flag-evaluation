# terraform {
#   required_providers {
#     ldflags = {
#       version = "0.2"
#       source  = "launchdarkly/ldflags"
#     }
#     kubernetes = {
#       version = "2.11.0"
#       source  = "hashicorp/kubernetes"
#     }
#   }
# }

# provider "kubernetes" {
#   config_path    = "~/.kube/config"
#   config_context = "minikube"
# }

# provider "ldflags" {
#   host = ""
#   sdk_key = ""
# }
