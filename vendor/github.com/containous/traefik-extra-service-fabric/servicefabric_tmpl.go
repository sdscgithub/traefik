package servicefabric

const tmpl = `
[backends]
{{range $aggName, $aggServices := getGroupedServices .Services }}
  [backends."{{ $aggName }}"]
  {{range $service := $aggServices }}
  {{range $partition := $service.Partitions }}
  {{range $instance := $partition.Instances }}
    [backends."{{ $aggName }}".servers."{{ $service.ID }}-{{ $instance.ID }}"]
      url = "{{ getDefaultEndpoint $instance }}"
      weight = {{ getGroupedWeight $service }}
  {{end}}
  {{end}}
  {{end}}
{{end}}

{{range $service := .Services }}
  {{if isEnabled $service }}
    {{range $partition := $service.Partitions }}

      {{if isStateless $service }}

        {{ $backendName := $service.Name }}
        [backends."{{ $backendName }}"]

        {{ $circuitBreaker := getCircuitBreaker $service }}
        {{if $circuitBreaker }}
          [backends."{{ $backendName }}".circuitBreaker]
            expression = "{{ $circuitBreaker.Expression }}"
        {{end}}

        {{ $loadBalancer := getLoadBalancer $service }}
        {{if $loadBalancer }}
          [backends."{{ $backendName }}".loadBalancer]
            method = "{{ $loadBalancer.Method }}"
            sticky = {{ $loadBalancer.Sticky }}
            {{if $loadBalancer.Stickiness }}
            [backends."{{ $backendName }}".loadBalancer.stickiness]
              cookieName = "{{ $loadBalancer.Stickiness.CookieName }}"
            {{end}}
        {{end}}

        {{ $maxConn := getMaxConn $service }}
        {{if $maxConn }}
          [backends."{{ $backendName }}".maxConn]
            extractorFunc = "{{ $maxConn.ExtractorFunc }}"
            amount = {{ $maxConn.Amount }}
        {{end}}

        {{ $healthCheck := getHealthCheck $service }}
        {{if $healthCheck }}
          [backends."{{ $backendName }}".healthCheck]
            path = "{{ $healthCheck.Path }}"
            port = {{ $healthCheck.Port }}
            interval = "{{ $healthCheck.Interval }}"
        {{end}}

        {{range $instance := $partition.Instances}}
          [backends."{{ $service.Name }}".servers."{{ $instance.ID }}"]
            url = "{{ getDefaultEndpoint $instance }}"
            weight = {{ getLabelValue $service "backend.weight" "1" }}
        {{end}}

      {{else if isStateful $service}}

        {{range $replica := $partition.Replicas}}
          {{if isPrimary $replica}}
            {{ $backendName := getBackendName $service $partition }}
            [backends."{{ $backendName }}".servers."{{ $replica.ID }}"]
              url = "{{ getDefaultEndpoint $replica }}"
							weight = 1

            [backends."{{$backendName}}".LoadBalancer]
              method = "drr"

          {{end}}
        {{end}}

      {{end}}

    {{end}}
  {{end}}
{{end}}

[frontends]
{{range $groupName, $groupServices := getGroupedServices .Services }}
  {{ $service := index $groupServices 0 }}
  [frontends."{{ $groupName }}"]
    backend = "{{ $groupName }}"
    priority = 50

  {{range $key, $value := getFrontendRules $service }}
    [frontends."{{ $groupName }}".routes."{{ $key }}"]
      rule = "{{ $value }}"
  {{end}}
{{end}}

{{range $service := .Services }}
  {{if isEnabled $service }}
    {{ $frontendName := $service.Name }}

    {{if isStateless $service }}

      [frontends."frontend-{{ $frontendName }}"]
        backend = "{{ $service.Name }}"
        passHostHeader = {{ getPassHostHeader $service }}
        passTLSCert = {{ getPassTLSCert $service }}
        priority = {{ getPriority $service }}
  
        {{ $entryPoints := getEntryPoints $service }}
        {{if $entryPoints }}
        entryPoints = [{{range $entryPoints }}
          "{{.}}",
          {{end}}]
        {{end}}
  
        {{ $basicAuth := getBasicAuth $service }}
        {{if $basicAuth }}
         basicAuth = [{{range $basicAuth }}
          "{{.}}",
          {{end}}]
        {{end}}

        {{ $whitelist := getWhiteList $service }}
        {{if $whitelist }}
        [frontends."frontend-{{ $frontendName }}".whiteList]
          sourceRange = [{{range $whitelist.SourceRange }}
            "{{.}}",
            {{end}}]
          useXForwardedFor = {{ $whitelist.UseXForwardedFor }}
        {{end}}

        {{ $redirect := getRedirect $service }}
        {{if $redirect }}
        [frontends."frontend-{{ $frontendName }}".redirect]
          entryPoint = "{{ $redirect.EntryPoint }}"
          regex = "{{ $redirect.Regex }}"
          replacement = "{{ $redirect.Replacement }}"
          permanent = {{ $redirect.Permanent }}
        {{end}}

        {{ $headers := getHeaders $service }}
        {{if $headers }}
        [frontends."frontend-{{ $frontendName }}".headers]
          SSLRedirect = {{ $headers.SSLRedirect }}
          SSLTemporaryRedirect = {{ $headers.SSLTemporaryRedirect }}
          SSLHost = "{{ $headers.SSLHost }}"
          STSSeconds = {{ $headers.STSSeconds }}
          STSIncludeSubdomains = {{ $headers.STSIncludeSubdomains }}
          STSPreload = {{ $headers.STSPreload }}
          ForceSTSHeader = {{ $headers.ForceSTSHeader }}
          FrameDeny = {{ $headers.FrameDeny }}
          CustomFrameOptionsValue = "{{ $headers.CustomFrameOptionsValue }}"
          ContentTypeNosniff = {{ $headers.ContentTypeNosniff }}
          BrowserXSSFilter = {{ $headers.BrowserXSSFilter }}
          CustomBrowserXSSValue = "{{ $headers.CustomBrowserXSSValue }}"
          ContentSecurityPolicy = "{{ $headers.ContentSecurityPolicy }}"
          PublicKey = "{{ $headers.PublicKey }}"
          ReferrerPolicy = "{{ $headers.ReferrerPolicy }}"
          IsDevelopment = {{ $headers.IsDevelopment }}
  
          {{if $headers.AllowedHosts }}
          AllowedHosts = [{{range $headers.AllowedHosts }}
            "{{.}}",
            {{end}}]
          {{end}}
  
          {{if $headers.HostsProxyHeaders }}
          HostsProxyHeaders = [{{range $headers.HostsProxyHeaders }}
            "{{.}}",
            {{end}}]
          {{end}}
  
          {{if $headers.CustomRequestHeaders }}
          [frontends."frontend-{{ $frontendName }}".headers.customRequestHeaders]
            {{range $k, $v := $headers.CustomRequestHeaders }}
            {{$k}} = "{{$v}}"
            {{end}}
          {{end}}
  
          {{if $headers.CustomResponseHeaders }}
          [frontends."frontend-{{ $frontendName }}".headers.customResponseHeaders]
            {{range $k, $v := $headers.CustomResponseHeaders }}
            {{$k}} = "{{$v}}"
            {{end}}
          {{end}}
  
          {{if $headers.SSLProxyHeaders }}
          [frontends."frontend-{{ $frontendName }}".headers.SSLProxyHeaders]
            {{range $k, $v := $headers.SSLProxyHeaders }}
            {{$k}} = "{{$v}}"
            {{end}}
          {{end}}
        {{end}}
  
      {{range $key, $value := getFrontendRules $service }}
        [frontends."frontend-{{ $frontendName }}".routes."{{ $key }}"]
          rule = "{{ $value }}"
      {{end}}

    {{else if isStateful $service}}

      {{range $partition := $service.Partitions }}
        {{ $partitionId := $partition.PartitionInformation.ID }}

        {{if hasLabel $service "frontend.rule" }}
          [frontends."{{ $service.Name }}/{{ $partitionId }}"]
            backend = "{{ getBackendName $service.Name $partition }}"

          [frontends."{{ $service.Name }}/{{ $partitionId }}".routes.default]
            rule = {{ getLabelValue $service "frontend.rule.partition.$partitionId" "" }}
        {{end}}
      {{end}}

    {{end}}

  {{end}}
{{end}}
`
