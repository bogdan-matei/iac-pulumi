#!/bin/bash
set -o pipefail

# Note: script requires sudo


# Install and configure Docker
docker_install (){
	apt-get update -y;

	apt-get install -y \
			ca-certificates \
			curl \
			gnupg \
			lsb-release;

	curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg;

	echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
		$(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null;

	apt-get update -y;

	# install will creash if the override.conf exists in /etc/...
	apt-get install -y docker-ce docker-ce-cli containerd.io;

	
	# Set Driver cgroup driver to systemd (match kubeadm)
	# due to https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/#installing-runtime


	echo "Changing cgroupdriver of docker to 'systemd' ..."
	systemctl stop docker;
	mkdir -p /etc/systemd/system/docker.service.d && touch /etc/systemd/system/docker.service.d/override.conf
	echo "ExecStart=/usr/bin/dockerd --exec-opt native.cgroupdriver=systemd" | tee /etc/systemd/system/docker.service.d/override.conf
	systemctl daemon-reload;
	systemctl restart docker;

	docker run hello-world;

	docker_check=$?
	if [ $docker_check -ne 0 ]; then
		echo "Docker run can't run. Exit with status from command $docker_check.";
		return 1;
	fi

	return 0;
}

kube_components (){
	
	# Check https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/
	# Install kubeadm, kubelet and kubectl

	apt-get update -y;
	apt-get install -y apt-transport-https ca-certificates curl -y;

	sudo curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg;

	echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list

	apt-get update;
	apt-get install -y kubelet kubeadm kubectl;
	apt-mark hold kubelet kubeadm kubectl;


	# Install CNI plugin (Calico)
	kubectl apply -f https://docs.projectcalico.org/manifests/calico.yaml

	if [ $(kubectl get pods -n kube-system | grep calico -c) -eq 0 ]; then
		echo "Calico wasn't installed. Manually troubleshoot the issue..."
		return 1;
	fi

	return 0;
}

init_node() {
	if [ $(docker_install) -ne 0 ]; then
		echo "Docker install failed";
		exit -1;
	fi

	if [ $(kube_components) -ne 0 ]; then
		echo "Kubernetes components install failed";
		exit -1;
	fi

	exit 0;
}

# Join node, for details check https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/#join-nodes
node_join(){
	:'
	$1 - <token> value
	$2 - <control-plane-host>:<control-plane-port> as a string
	$3 - <hash> of the token
	'
	kubeadm join --token $1 $2 --discovery-token-ca-cert-hash sha256:$3;

	exit 0;
}