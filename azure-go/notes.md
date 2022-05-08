# pages:
- https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/
- https://docs.docker.com/engine/install/ubuntu/

# notes

sudo systemctl edit docker -> add docker with systemd cgroup

[Service]
ExecStart=/usr/bin/dockerd --exec-opt native.cgroupdriver=systemd

- Finished installing https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/#pod-network 'kubectl describe node clustercontrolplane71637dd2' comands runs and shows that cni and kubelet is rdy 

- to use externally, an additional configuration with the hostname is required: https://blog.scottlowe.org/2019/07/30/adding-a-name-to-kubernetes-api-server-certificate/
  
# playing

  *Warning  FailedScheduling  38s   default-scheduler  0/1 nodes are available: 1 node(s) had taint {node-role.kubernetes.io/master: }, that the pod didn't tolerate.* for  kubectl apply -f https://k8s.io/examples/application/deployment.yaml

	Useful comand to check cgroup of docker: docker system info | grep -i drive

	Token expired, and can be regenerated https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/#join-nodes

kubeadm join 10.0.2.4:6443 --token $token --discovery-token-ca-cert-hash "sha256:$sha" 

	For kubeconfig:
	- create SA -> kubectl create serviceaccount admin -n kube-system
	- bind it to a cluster role -> kubectl create clusterrolebinding admin --clusterrole=cluster-admin --serviceaccount=kube-system:admin
	- create kubeconfig file
  	- use info from `kubectl config view --flatten --minify` and update the cluster,context and user; update your local file afterwards 