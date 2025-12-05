resource "kubernetes_namespace" "devops-staging" {
  metadata {
    name = "test"
    labels = {
      app = "test"
    }
  }
}