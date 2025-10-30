resource "kubernetes_namespace" "devops-staging" {
  metadata {
    name = "devops-staging"
    labels = {
      app = "devops-staging"
    }
  }
}