install-k3d:
	curl -s https://raw.githubusercontent.com/rancher/k3d/master/install.sh | bash
	## porter mixin install helm3 --version v0.1.0 --url  https://github.com/MChorfa/porter-helm3/releases/download
deploy-cluster:
	# Create
	k3d create \
	--name="mccluster" \
	--workers="1" \
	--publish="80:80" \
	--publish="443:443"
	# Export config
	export KUBECONFIG=$(k3d get-kubeconfig --name='mccluster')
	# Get cluster info
	kubectl cluster-info
	# Watch all pods
	kubectl get pods --all-namespaces -w

destroy-cluster:
	k3d delete --name mccluster

install-porter:
	kubectl config use-context mccluster
	porter build --debug
	porter creds generate
	porter install \
	--cred kubeconfig \
	--param POSTGRESQL_DATABASE="password123" \
	--param POSTGRESQL_WORKER_USERNAME="password123" \
	--param POSTGRESQL_WORKER_PASSWORD="password123" \
	--param POSTGRESQL_ADMIN_PASSWORD="password123" \
	--param KEYCLOAK_USERNAME="password123" \
	--param KEYCLOAK_PASSWORD="password123" --debug

uninstall-porter:
	porter uninstall \
	--cred kubeconfig \
	--param POSTGRESQL_DATABASE="password123" \
	--param POSTGRESQL_WORKER_USERNAME="password123" \
	--param POSTGRESQL_WORKER_PASSWORD="password123" \
	--param POSTGRESQL_ADMIN_PASSWORD="password123" \
	--param KEYCLOAK_USERNAME="password123" \
	--param KEYCLOAK_PASSWORD="password123"

