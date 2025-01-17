package ovn

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/config"
	ovntest "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/testing"
	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/types"
	"github.com/urfave/cli/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachinerytypes "k8s.io/apimachinery/pkg/types"
)

type service struct{}

func newServiceMeta(name, namespace string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		UID:       apimachinerytypes.UID(namespace),
		Name:      name,
		Namespace: namespace,
		Labels: map[string]string{
			"name": name,
		},
	}
}

func newService(name, namespace, ip string, ports []v1.ServicePort, serviceType v1.ServiceType, externalIPs []string) *v1.Service {
	return &v1.Service{
		ObjectMeta: newServiceMeta(name, namespace),
		Spec: v1.ServiceSpec{
			ClusterIP:   ip,
			Ports:       ports,
			Type:        serviceType,
			ExternalIPs: externalIPs,
		},
	}
}

func (s service) baseCmds(fexec *ovntest.FakeExec, service v1.Service) {
	fexec.AddFakeCmd(&ovntest.ExpectedCmd{
		Cmd:    "ovn-nbctl --timeout=15 --data=bare --no-heading --columns=_uuid find load_balancer external_ids:k8s-cluster-lb-tcp=yes",
		Output: k8sTCPLoadBalancerIP,
	})
	fexec.AddFakeCmdsNoOutputNoError([]string{
		"ovn-nbctl --timeout=15 --columns=name,_uuid --format=json find acl action=reject",
	})
	fexec.AddFakeCmd(&ovntest.ExpectedCmd{
		Cmd:    fmt.Sprintf("ovn-nbctl --timeout=15 --data=bare --no-heading get load_balancer %s vips", k8sTCPLoadBalancerIP),
		Output: "{\"172.30.0.10:53\"=\"10.128.0.18:5353,10.129.0.3:5353\"}",
	})
	fexec.AddFakeCmdsNoOutputNoError([]string{
		fmt.Sprintf("ovn-nbctl --timeout=15 --if-exists remove load_balancer %s vips \"172.30.0.10:53\"", k8sTCPLoadBalancerIP),
		fmt.Sprintf("ovn-nbctl --timeout=15 --data=bare --no-heading --columns=_uuid find acl name=%s-172.30.0.10\\:53", k8sTCPLoadBalancerIP),
	})
	fexec.AddFakeCmd(&ovntest.ExpectedCmd{
		Cmd:    "ovn-nbctl --timeout=15 --data=bare --no-heading --columns=_uuid find load_balancer external_ids:k8s-cluster-lb-udp=yes",
		Output: k8sUDPLoadBalancerIP,
	})
	fexec.AddFakeCmd(&ovntest.ExpectedCmd{
		Cmd:    fmt.Sprintf("ovn-nbctl --timeout=15 --data=bare --no-heading get load_balancer %s vips", k8sUDPLoadBalancerIP),
		Output: "{\"172.30.0.10:53\"=\"10.128.0.18:5353,10.129.0.3:5353\"}",
	})
	fexec.AddFakeCmdsNoOutputNoError([]string{
		fmt.Sprintf("ovn-nbctl --timeout=15 --if-exists remove load_balancer %s vips \"172.30.0.10:53\"", k8sUDPLoadBalancerIP),
		fmt.Sprintf("ovn-nbctl --timeout=15 --data=bare --no-heading --columns=_uuid find acl name=%s-172.30.0.10\\:53", k8sUDPLoadBalancerIP),
	})
	fexec.AddFakeCmd(&ovntest.ExpectedCmd{
		Cmd:    "ovn-nbctl --timeout=15 --data=bare --no-heading --columns=_uuid find load_balancer external_ids:k8s-cluster-lb-sctp=yes",
		Output: k8sSCTPLoadBalancerIP,
	})
	fexec.AddFakeCmd(&ovntest.ExpectedCmd{
		Cmd:    fmt.Sprintf("ovn-nbctl --timeout=15 --data=bare --no-heading get load_balancer %s vips", k8sSCTPLoadBalancerIP),
		Output: "{\"172.30.0.10:53\"=\"10.128.0.18:5353,10.129.0.3:5353\"}",
	})
	fexec.AddFakeCmdsNoOutputNoError([]string{
		fmt.Sprintf("ovn-nbctl --timeout=15 --if-exists remove load_balancer %s vips \"172.30.0.10:53\"", k8sSCTPLoadBalancerIP),
		fmt.Sprintf("ovn-nbctl --timeout=15 --data=bare --no-heading --columns=_uuid find acl name=%s-172.30.0.10\\:53", k8sSCTPLoadBalancerIP),
	})
	fexec.AddFakeCmd(&ovntest.ExpectedCmd{
		Cmd:    "ovn-nbctl --timeout=15 --data=bare --no-heading --columns=name find logical_router options:chassis!=null",
		Output: "gateway1",
	})
	fexec.AddFakeCmd(&ovntest.ExpectedCmd{
		Cmd:    "ovn-nbctl --timeout=15 --data=bare --no-heading --columns=_uuid find load_balancer external_ids:TCP_lb_gateway_router=gateway1",
		Output: "tcp_load_balancer_id_1",
	})
	fexec.AddFakeCmd(&ovntest.ExpectedCmd{
		Cmd:    "ovn-nbctl --timeout=15 --data=bare --no-heading get load_balancer tcp_load_balancer_id_1 vips",
		Output: "{\"172.30.0.10:53\"=\"10.128.0.18:5353,10.129.0.3:5353\"}",
	})
	fexec.AddFakeCmdsNoOutputNoError([]string{
		"ovn-nbctl --timeout=15 --if-exists remove load_balancer tcp_load_balancer_id_1 vips \"172.30.0.10:53\"",
		"ovn-nbctl --timeout=15 --data=bare --no-heading --columns=_uuid find acl name=tcp_load_balancer_id_1-172.30.0.10\\:53",
	})
	fexec.AddFakeCmd(&ovntest.ExpectedCmd{
		Cmd:    "ovn-nbctl --timeout=15 --data=bare --no-heading --columns=_uuid find load_balancer external_ids:UDP_lb_gateway_router=gateway1",
		Output: "udp_load_balancer_id_1",
	})
	fexec.AddFakeCmd(&ovntest.ExpectedCmd{
		Cmd:    "ovn-nbctl --timeout=15 --data=bare --no-heading get load_balancer udp_load_balancer_id_1 vips",
		Output: "{\"172.30.0.10:53\"=\"10.128.0.18:5353,10.129.0.3:5353\"}",
	})
	fexec.AddFakeCmdsNoOutputNoError([]string{
		"ovn-nbctl --timeout=15 --if-exists remove load_balancer udp_load_balancer_id_1 vips \"172.30.0.10:53\"",
		"ovn-nbctl --timeout=15 --data=bare --no-heading --columns=_uuid find acl name=udp_load_balancer_id_1-172.30.0.10\\:53",
	})
	fexec.AddFakeCmd(&ovntest.ExpectedCmd{
		Cmd:    "ovn-nbctl --timeout=15 --data=bare --no-heading --columns=_uuid find load_balancer external_ids:SCTP_lb_gateway_router=gateway1",
		Output: "sctp_load_balancer_id_1",
	})
	fexec.AddFakeCmd(&ovntest.ExpectedCmd{
		Cmd:    "ovn-nbctl --timeout=15 --data=bare --no-heading get load_balancer sctp_load_balancer_id_1 vips",
		Output: "{\"172.30.0.10:53\"=\"10.128.0.18:5353,10.129.0.3:5353\"}",
	})
	fexec.AddFakeCmdsNoOutputNoError([]string{
		"ovn-nbctl --timeout=15 --if-exists remove load_balancer sctp_load_balancer_id_1 vips \"172.30.0.10:53\"",
		"ovn-nbctl --timeout=15 --data=bare --no-heading --columns=_uuid find acl name=sctp_load_balancer_id_1-172.30.0.10\\:53",
		"ovn-nbctl --timeout=15 --data=bare --no-heading --columns=name find logical_router options:chassis!=null",
	})
	fexec.AddFakeCmd(&ovntest.ExpectedCmd{
		Cmd:    fmt.Sprintf("ovn-nbctl --timeout=15 --data=bare --no-heading --columns=_uuid find logical_switch load_balancer{>=}%s", k8sTCPLoadBalancerIP),
		Output: "62c672a4-1132-44ab-9202-e47d18784138",
	})
	fexec.AddFakeCmdsNoOutputNoError([]string{
		fmt.Sprintf("ovn-nbctl --timeout=15 --data=bare --no-heading --columns=name find logical_router load_balancer{>=}%s", k8sTCPLoadBalancerIP),
	})
}

func (s service) addCmds(fexec *ovntest.FakeExec, service v1.Service) {
	s.baseCmds(fexec, service)
	for _, port := range service.Spec.Ports {
		fexec.AddFakeCmdsNoOutputNoError([]string{
			fmt.Sprintf("ovn-nbctl --timeout=15 --data=bare --no-heading --columns=_uuid find acl name=%s-%s\\:%v",
				k8sTCPLoadBalancerIP, service.Spec.ClusterIP, port.Port),
			fmt.Sprintf("ovn-nbctl --timeout=15 --id=@reject-acl create acl direction="+types.DirectionFromLPort+" priority="+types.DefaultDenyPriority+" match=\"ip4.dst==%s && tcp "+
				"&& tcp.dst==%v\" action=reject log=false severity=info meter=acl-logging name=%s-%s\\:%v -- add port_group %s acls @reject-acl", service.Spec.ClusterIP, port.Port,
				k8sTCPLoadBalancerIP, service.Spec.ClusterIP, port.Port, ovnClusterPortGroupUUID),
		})
	}
}

func (s service) delCmds(fexec *ovntest.FakeExec, service v1.Service) {
	for _, port := range service.Spec.Ports {
		fexec.AddFakeCmdsNoOutputNoError([]string{
			fmt.Sprintf("ovn-nbctl --timeout=15 --if-exists remove load_balancer %s vips \"%s:%v\"", k8sTCPLoadBalancerIP, service.Spec.ClusterIP, port.Port),
			fmt.Sprintf("ovn-nbctl --timeout=15 --data=bare --no-heading --columns=_uuid find acl name=%s-%s\\:%v", k8sTCPLoadBalancerIP, service.Spec.ClusterIP, port.Port),
			"ovn-nbctl --timeout=15 --data=bare --no-heading --columns=name find logical_router options:chassis!=null",
			"ovn-nbctl --timeout=15 --data=bare --no-heading --columns=name find logical_router options:chassis!=null",
		})
	}
}

var _ = ginkgo.Describe("OVN Namespace Operations", func() {
	var (
		app     *cli.App
		fakeOvn *FakeOVN
		fExec   *ovntest.FakeExec
	)

	ginkgo.BeforeEach(func() {
		// Restore global default values before each testcase
		config.PrepareTestConfig()

		app = cli.NewApp()
		app.Name = "test"
		app.Flags = config.Flags

		fExec = ovntest.NewFakeExec()
		fakeOvn = NewFakeOVN(fExec)
	})

	ginkgo.AfterEach(func() {
		fakeOvn.shutdown()
	})

	ginkgo.Context("on startup", func() {

		ginkgo.It("reconciles an existing service", func() {
			app.Action = func(ctx *cli.Context) error {

				test := service{}

				service := *newService("service1", "namespace1", "10.129.0.2",
					[]v1.ServicePort{
						{
							Port:     8032,
							Protocol: v1.ProtocolTCP,
						},
					},
					v1.ServiceTypeClusterIP,
					nil,
				)

				test.addCmds(fExec, service)

				fakeOvn.start(ctx,
					&v1.ServiceList{
						Items: []v1.Service{
							service,
						},
					},
				)
				fakeOvn.controller.clusterPortGroupUUID = ovnClusterPortGroupUUID
				fakeOvn.controller.WatchServices()

				_, err := fakeOvn.fakeClient.KubeClient.CoreV1().Services(service.Namespace).Get(context.TODO(), service.Name, metav1.GetOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(fExec.CalledMatchesExpected()).To(gomega.BeTrue(), fExec.ErrorDesc)

				return nil
			}

			err := app.Run([]string{app.Name})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("reconciles a deleted service", func() {
			app.Action = func(ctx *cli.Context) error {

				test := service{}

				service := *newService("service1", "namespace1", "10.129.0.2",
					[]v1.ServicePort{
						{
							Port:     8032,
							Protocol: v1.ProtocolTCP,
						},
					},
					v1.ServiceTypeClusterIP,
					nil,
				)

				test.addCmds(fExec, service)

				fakeOvn.start(ctx,
					&v1.ServiceList{
						Items: []v1.Service{
							service,
						},
					},
				)
				fakeOvn.controller.clusterPortGroupUUID = ovnClusterPortGroupUUID
				fakeOvn.controller.WatchServices()

				_, err := fakeOvn.fakeClient.KubeClient.CoreV1().Services(service.Namespace).Get(context.TODO(), service.Name, metav1.GetOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(fExec.CalledMatchesExpected()).To(gomega.BeTrue(), fExec.ErrorDesc)

				test.delCmds(fExec, service)
				err = fakeOvn.fakeClient.KubeClient.CoreV1().Services(service.Namespace).Delete(context.TODO(), service.Name, *metav1.NewDeleteOptions(0))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Eventually(fExec.CalledMatchesExpected).Should(gomega.BeTrue(), fExec.ErrorDesc)

				s, err := fakeOvn.fakeClient.KubeClient.CoreV1().Services(service.Namespace).Get(context.TODO(), service.Name, metav1.GetOptions{})
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(s).To(gomega.BeNil())

				return nil
			}

			err := app.Run([]string{app.Name})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})
	})
})
