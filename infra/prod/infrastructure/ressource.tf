data "kubernetes_namespace" "prod" {
  metadata {
    name = "prod"
    labels = {
      app = "prod"
    }
  }
}