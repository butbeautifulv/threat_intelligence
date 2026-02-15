#CloudProvider: "aws" | "gcp" | "azure"

#TerraformNetwork: {
    vpc_cidr: string | *"10.0.0.0/16"
    private_subnets: [...string] | *["10.0.1.0/24", "10.0.2.0/24"]
    public_subnets:  [...string] | *[]
    // по умолчанию — только приватные подсети
}

#TerraformSecurityGroupRule: {
    description: string
    protocol: "tcp" | "udp" | "icmp" | *"tcp"
    from_port: int
    to_port:   int
    cidr_blocks: [...string]
}

#TerraformKMS: {
    enabled: bool | *true
    key_alias: string | *"alias/app-default"
}

#TerraformIAMPolicy: {
    // Явно перечисленные действия, без wildcard
    actions: [...string] & ![=~".*:.*"] // запрет "*:*"
    resources: [...string]
}

#TerraformLogging: {
    enabled: bool | *true
    retention_days: int | *30
}

#TerraformModuleSecurity: {
    encryption_at_rest: bool | *true
    encryption_in_transit: bool | *true

    kms: #TerraformKMS

    logging: #TerraformLogging

    // Сетевые правила по умолчанию — deny all, разрешаем только явно
    security_group_ingress: [...#TerraformSecurityGroupRule] | *[]
    security_group_egress: [...#TerraformSecurityGroupRule] | *[{
        description: "Allow egress to internet only if explicitly needed"
        protocol:    "tcp"
        from_port:   443
        to_port:     443
        cidr_blocks: ["0.0.0.0/0"]
    }]
}

#TerraformModuleInputs: {
    name: string
    environment: "dev" | "stage" | "prod" | *"dev"

    cloud: #CloudProvider

    network: #TerraformNetwork

    security: #TerraformModuleSecurity

    // Теги/labels для аудита и DevSecOps
    tags: {
        "owner":        string
        "environment":  string
        "cost-center"?: string
        "data-classification"?: "public" | "internal" | "confidential" | *"internal"
    }

    // Опциональные параметры для конкретного провайдера
    aws?: {
        region: string | *"eu-central-1"
        // запрет на публичные IP по умолчанию
        associate_public_ip_address: bool | *false
    }

    gcp?: {
        region: string | *"europe-west3"
        // private Google access
        private_google_access: bool | *true
    }

    azure?: {
        location: string | *"westeurope"
        // private endpoints по умолчанию
        private_endpoint_enabled: bool | *true
    }
}
