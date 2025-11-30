resource "helm_release" "argocd" {
  name       = "argocd"
  repository = "https://argoproj.github.io/argo-helm"
  chart      = "argo-cd"
  namespace  = data.kubernetes_namespace.prod.metadata[0].name
  version    = var.argocd_chart_version
  values     = [file("${path.module}/argoCD/argoCD-values.yaml")]

  create_namespace = false

  depends_on = [data.kubernetes_namespace.prod]
}

resource "helm_release" "argo_rollouts" {
  name       = "argo-rollouts"
  repository = "https://argoproj.github.io/argo-helm"
  chart      = "argo-rollouts"
  namespace  = data.kubernetes_namespace.prod.metadata[0].name
  version    = var.argo_rollouts_chart_version
  create_namespace = false
  depends_on = [data.kubernetes_namespace.prod]
}