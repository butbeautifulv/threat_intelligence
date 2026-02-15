// Безопасные значения для Helm values

#HelmSecurityContext: {
    runAsNonRoot:   bool | *true
    runAsUser:      int  | *1000
    runAsGroup:     int  | *1000
    readOnlyRootFilesystem: bool | *true
    allowPrivilegeEscalation: bool | *false
    privileged: bool | *false
    capabilities: {
        drop: [...string] | *["ALL"]
        add?: [...string] // только если очень нужно
    }
}

#HelmPodSecurityContext: {
    fsGroup:        int  | *1000
    runAsNonRoot:   bool | *true
    seccompProfile: {
        type: "RuntimeDefault" | *"RuntimeDefault"
    }
}

#HelmResources: {
    limits: {
        cpu:    string | *"500m"
        memory: string | *"512Mi"
    }
    requests: {
        cpu:    string | *"100m"
        memory: string | *"128Mi"
    }
}

#HelmImage: {
    repository: string
    tag:        string
    pullPolicy: "IfNotPresent" | "Always" | *"IfNotPresent"
}

#HelmService: {
    type: "ClusterIP" | "NodePort" | "LoadBalancer" | *"ClusterIP"
    port: int | *80
}

#HelmIngress: {
    enabled: bool | *false
    className?: string
    hosts?: [...{
        host: string
        paths: [...{
            path:     string | *"/"
            pathType: "Prefix" | "Exact" | *"Prefix"
        }]
    }]
    tls?: [...{
        hosts: [...string]
        secretName: string
    }]
}

#HelmNetworkPolicy: {
    enabled: bool | *true
    ingress: [...{
        fromNamespaces?: [...string] // whitelist
        fromPods?: [...string]
        ports: [...{
            port: int
            protocol: "TCP" | "UDP" | *"TCP"
        }]
    }] | *[]
    egress: [...{
        toCIDRs?: [...string]
        toNamespaces?: [...string]
        ports: [...{
            port: int
            protocol: "TCP" | "UDP" | *"TCP"
        }]
    }] | *[]
}

#HelmProbes: {
    livenessProbe: {
        httpGet?: {
            path: string | *"/healthz"
            port: int    | *8080
        }
        initialDelaySeconds: int | *10
        periodSeconds:       int | *10
    }
    readinessProbe: {
        httpGet?: {
            path: string | *"/readyz"
            port: int    | *8080
        }
        initialDelaySeconds: int | *5
        periodSeconds:       int | *5
    }
}

#HelmSecretsRef: {
    enabled: bool | *true
    secretName: string
    // только ссылки, без явных секретов
    envFrom?: bool | *true
}

#HelmChartValues: {
    nameOverride?: string
    fullnameOverride?: string

    image: #HelmImage

    replicaCount: int | *2

    resources: #HelmResources

    securityContext:     #HelmSecurityContext
    podSecurityContext:  #HelmPodSecurityContext

    service: #HelmService
    ingress: #HelmIngress

    networkPolicy: #HelmNetworkPolicy

    probes: #HelmProbes

    secrets: #HelmSecretsRef

    // Дополнительные аннотации/labels для DevSecOps
    podAnnotations?: {
        "seccomp.security.alpha.kubernetes.io/pod"?: string
        "container.apparmor.security.beta.kubernetes.io/main"?: string
    }

    podLabels?: {
        "app.kubernetes.io/part-of"?: string
        "security-tier"?: "public" | "internal" | "restricted" | *"internal"
    }

    // Опциональный PodDisruptionBudget
    pdb?: {
        enabled: bool | *true
        minAvailable: string | *"50%"
    }

    // Опциональный HPA
    autoscaling?: {
        enabled: bool | *false
        minReplicas: int | *2
        maxReplicas: int | *5
        targetCPUUtilizationPercentage: int | *70
    }
}
