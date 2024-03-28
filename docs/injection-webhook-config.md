# Configuring access to the injection webhook

By default, the webhook is "opt-in" and requires a label to be applied to the namespace: `images.stvz.io/inject: "true"`.  This is configurable based on your individual requirements and you can modify the selectors and conditions in the [config/webhook/kustomization.yaml](kustomization configuration) file.


Some possiblities could include:

### Using namespace labels to manage access

If you use labels to identify namespace usage, you can use namespace selectors. Given a namespace with the following (this is the default matcher for requests):

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: myspace
  labels:
    - images.stvz.io/inject: "true"
```

You could replace the default selectors with this:

```yaml
namespaceSelector:
  matchExpression:
    - key: images.stvz.io/inject
      operator: In
      values:
        - "true"
```

This will send all requests for the supported/managed objects in the labeled namespace to the injector.

The opposite, could include including all namespaces except labeled namespaces.  If you have system namespaces that include resources that you do not want to handle with the webhook, label them in much the same way that was previously done eg. (`images.stvz.io/no-inject: "true"`) and add the following match rule:

```yaml
namespaceSelector:
  matchExpression:
    - key: images.stvz.io/no-inject
      operator: NotIn
      values:
        - "true"
```

Alternatively if you want to limit coral webhook access to you production and staging namespaces (given that the labels exist), while keeping development environments more flexible you can replace the default conditions with the following:

```yaml
namespaceSelector
  matchExpressions:
  - key: env
    operator: In
    values:
    - staging
    - production
```

In later versions of Kubernetes (1.28+) a new matching field was introduced called `matchConditions` which allows CEL style expressions with additional flexibility to query the request, authorizers, and objects directly.  As an example, you can forgo namespace labeling and specify namespaces to include and exclude directly.

```yaml
matchConditions:
  - name: exclude-namespaces
    expression: '!(object.metadata.namespace in ["kube-*", "*-system", "coral", "cert-manager"])'
```