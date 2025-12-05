variable "kubeconfig_path" {
  type    = string
  default = "./.kube/config"
  description = "Chemin vers le kubeconfig à utiliser pour provider kubernetes/helm"
}