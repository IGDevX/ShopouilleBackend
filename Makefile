KUBE_NS ?= monitoring
RELEASE ?= monitoring
CHART_DIR := infra/monitoring

.PHONY: deps
deps:
	helm dependency update $(CHART_DIR)

.PHONY: install-dev
install-dev: deps
	helm upgrade --install $(RELEASE) $(CHART_DIR) -n $(KUBE_NS) -f $(CHART_DIR)/values.yaml -f $(CHART_DIR)/values-dev.yaml --create-namespace

.PHONY: install-staging
install-staging: deps
	helm upgrade --install $(RELEASE) $(CHART_DIR) -n $(KUBE_NS) -f $(CHART_DIR)/values.yaml -f $(CHART_DIR)/values-staging.yaml --create-namespace

.PHONY: install-prod
install-prod: deps
	helm upgrade --install $(RELEASE) $(CHART_DIR) -n $(KUBE_NS) -f $(CHART_DIR)/values.yaml -f $(CHART_DIR)/values-prod.yaml --create-namespace

.PHONY: uninstall
uninstall:
	helm uninstall $(RELEASE) -n $(KUBE_NS) || true

.PHONY: pf-grafana
pf-grafana:
	kubectl -n $(KUBE_NS) port-forward svc/$(RELEASE)-grafana 3000:80

.PHONY: pf-prom
pf-prom:
	kubectl -n $(KUBE_NS) port-forward svc/$(RELEASE)-kube-prometheus-sta-prometheus 9090:9090

.PHONY: pf-am
pf-am:
	kubectl -n $(KUBE_NS) port-forward svc/$(RELEASE)-alertmanager 9093:9093

.PHONY: status
status:
	kubectl get pods -n $(KUBE_NS)
	kubectl get servicemonitors,podmonitors -A | grep $(RELEASE) || true

.PHONY: test-targets
test-targets:
	kubectl -n $(KUBE_NS) exec deploy/$(RELEASE)-kube-prometheus-sta-operator -- \
	  wget -qO- http://$(RELEASE)-kube-prometheus-sta-prometheus:9090/api/v1/targets | jq '.data.activeTargets | length'

