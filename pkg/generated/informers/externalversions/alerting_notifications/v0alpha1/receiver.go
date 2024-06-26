// SPDX-License-Identifier: AGPL-3.0-only

// Code generated by informer-gen. DO NOT EDIT.

package v0alpha1

import (
	"context"
	time "time"

	alertingnotificationsv0alpha1 "github.com/grafana/grafana/pkg/apis/alerting_notifications/v0alpha1"
	versioned "github.com/grafana/grafana/pkg/generated/clientset/versioned"
	internalinterfaces "github.com/grafana/grafana/pkg/generated/informers/externalversions/internalinterfaces"
	v0alpha1 "github.com/grafana/grafana/pkg/generated/listers/alerting_notifications/v0alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// ReceiverInformer provides access to a shared informer and lister for
// Receivers.
type ReceiverInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v0alpha1.ReceiverLister
}

type receiverInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewReceiverInformer constructs a new informer for Receiver type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewReceiverInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredReceiverInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredReceiverInformer constructs a new informer for Receiver type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredReceiverInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NotificationsV0alpha1().Receivers(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NotificationsV0alpha1().Receivers(namespace).Watch(context.TODO(), options)
			},
		},
		&alertingnotificationsv0alpha1.Receiver{},
		resyncPeriod,
		indexers,
	)
}

func (f *receiverInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredReceiverInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *receiverInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&alertingnotificationsv0alpha1.Receiver{}, f.defaultInformer)
}

func (f *receiverInformer) Lister() v0alpha1.ReceiverLister {
	return v0alpha1.NewReceiverLister(f.Informer().GetIndexer())
}
