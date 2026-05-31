/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	// TODO (user): Add any additional imports if needed
)

var _ = Describe("Pod Webhook", func() {
	var (
		obj       *corev1.Pod
		oldObj    *corev1.Pod
		defaulter PodCustomDefaulter
	)

	BeforeEach(func() {
		obj = &corev1.Pod{}
		oldObj = &corev1.Pod{}
		defaulter = PodCustomDefaulter{}
		Expect(defaulter).NotTo(BeNil(), "Expected defaulter to be initialized")
		Expect(oldObj).NotTo(BeNil(), "Expected oldObj to be initialized")
		Expect(obj).NotTo(BeNil(), "Expected obj to be initialized")
	})

	AfterEach(func() {
		// TODO (user): Add any teardown logic common to all tests
	})

	Context("When creating Pod under Defaulting Webhook", func() {
		It("Should inject the manifestation scheduling gate", func() {
			Expect(defaulter.Default(ctx, obj)).To(Succeed())
			Expect(gateNames(obj)).To(ContainElement("manifestation.pez.sh/awaiting-manifestation"))
		})

		It("Should not inject the gate twice", func() {
			Expect(defaulter.Default(ctx, obj)).To(Succeed())
			Expect(defaulter.Default(ctx, obj)).To(Succeed())
			Expect(obj.Spec.SchedulingGates).To(HaveLen(1))
		})

		It("Should respect the skip label", func() {
			obj.Labels = map[string]string{"manifestation.pez.sh/skip": "true"}
			Expect(defaulter.Default(ctx, obj)).To(Succeed())
			Expect(gateNames(obj)).NotTo(ContainElement("manifestation.pez.sh/awaiting-manifestation"))
		})

		It("Should preserve pre-existing scheduling gates", func() {
			obj.Spec.SchedulingGates = []corev1.PodSchedulingGate{{Name: "example.com/other"}}
			Expect(defaulter.Default(ctx, obj)).To(Succeed())
			Expect(gateNames(obj)).To(ContainElements("example.com/other", "manifestation.pez.sh/awaiting-manifestation"))
		})
	})

})

func gateNames(pod *corev1.Pod) []string {
	names := make([]string, 0, len(pod.Spec.SchedulingGates))
	for _, g := range pod.Spec.SchedulingGates {
		names = append(names, g.Name)
	}
	return names
}
