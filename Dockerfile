FROM busybox
MAINTAINER Joel Meyer <joel.meyer@gmail.com>
ADD kubegateway kubegateway
ADD kubegateway.go kubegateway.go
ENTRYPOINT ["/kubegateway"]
