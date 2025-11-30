variable "kubeconfig_path" {
  type    = string
  default = "./.kube/config"
  description = "Chemin vers le kubeconfig à utiliser pour provider kubernetes/helm"
}

variable "argocd_chart_version" {
  type    = string
  default = "9.0.5"
}

variable "argo_rollouts_chart_version" {
  type    = string
  default = "2.40.5"
}