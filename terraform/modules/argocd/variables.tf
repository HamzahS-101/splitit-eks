variable "argocd_chart_version" {
  type        = string
  description = "Version of the argo-cd Helm chart."
}

variable "argocd_apps_chart_version" {
  type        = string
  description = "Version of the argocd-apps Helm chart that creates the root Application."
}

variable "argocd_values" {
  type        = string
  description = "Contents of the Argo CD Helm values file."
}

variable "root_app_values" {
  type        = string
  description = "Contents of the values file that defines the root Application."
}
