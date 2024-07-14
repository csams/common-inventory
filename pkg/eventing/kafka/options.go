package kafka

import (
	"strings"

	"github.com/spf13/pflag"
)

type Opt struct {
	Name        string
	Value       string
	Description string
}

type Options struct {
	Opts []Opt
}

func NewOptions() *Options {
	return &Options{
		Opts: []Opt{
			{"builtin.features", "gzip, snappy, ssl, sasl, regex, lz4, sasl_gssapi, sasl_plain, sasl_scram, plugins, zstd, sasl_oauthbearer, http, oidc", "Indicates the builtin features for this build of librdkafka. An application can either query this value or attempt to set it with its list of required features to check for library support. \n*Type: CSV flags*"},
			{"client.id", "rdkafka", "Client identifier. \n*Type: string*"},
			{"metadata.broker.list", "", "Initial list of brokers as a CSV list of broker host or host:port. The application may also use `rd_kafka_brokers_add()` to add brokers during runtime. \n*Type: string*"},
			{"bootstrap.servers", "", "Alias for `metadata.broker.list`: Initial list of brokers as a CSV list of broker host or host:port. The application may also use `rd_kafka_brokers_add()` to add brokers during runtime. \n*Type: string*"},
			{"message.max.bytes", "1000000", "Maximum Kafka protocol request message size. Due to differing framing overhead between protocol versions the producer is unable to reliably enforce a strict max message limit at produce time and may exceed the maximum size by one message in protocol ProduceRequests, the broker will enforce the the topic's `max.message.bytes` limit (see Apache Kafka documentation). \n*Type: integer*"},
			{"message.copy.max.bytes", "65535", "Maximum size for message to be copied to buffer. Messages larger than this will be passed by reference (zero-copy) at the expense of larger iovecs. \n*Type: integer*"},
			{"receive.message.max.bytes", "100000000", "Maximum Kafka protocol response message size. This serves as a safety precaution to avoid memory exhaustion in case of protocol hickups. This value must be at least `fetch.max.bytes`  + 512 to allow for protocol overhead; the value is adjusted automatically unless the configuration property is explicitly set. \n*Type: integer*"},
			{"max.in.flight.requests.per.connection", "1000000", "Maximum number of in-flight requests per broker connection. This is a generic property applied to all broker communication, however it is primarily relevant to produce requests. In particular, note that other mechanisms limit the number of outstanding consumer fetch request per broker to one. \n*Type: integer*"},
			{"max.in.flight", "1000000", "Alias for `max.in.flight.requests.per.connection`: Maximum number of in-flight requests per broker connection. This is a generic property applied to all broker communication, however it is primarily relevant to produce requests. In particular, note that other mechanisms limit the number of outstanding consumer fetch request per broker to one. \n*Type: integer*"},
			{"topic.metadata.refresh.interval.ms", "300000", "Period of time in milliseconds at which topic and broker metadata is refreshed in order to proactively discover any new brokers, topics, partitions or partition leader changes. Use -1 to disable the intervalled refresh (not recommended). If there are no locally referenced topics (no topic objects created, no messages produced, no subscription or no assignment) then only the broker list will be refreshed every interval but no more often than every 10s. \n*Type: integer*"},
			{"metadata.max.age.ms", "900000", "Metadata cache max age. Defaults to topic.metadata.refresh.interval.ms * 3 \n*Type: integer*"},
			{"topic.metadata.refresh.fast.interval.ms", "100", "When a topic loses its leader a new metadata request will be enqueued immediately and then with this initial interval, exponentially increasing upto `retry.backoff.max.ms`, until the topic metadata has been refreshed. If not set explicitly, it will be defaulted to `retry.backoff.ms`. This is used to recover quickly from transitioning leader brokers. \n*Type: integer*"},
			{"topic.metadata.refresh.fast.cnt", "10", "**DEPRECATED** No longer used. \n*Type: integer*"},
			{"topic.metadata.refresh.sparse", "true", "Sparse metadata requests (consumes less network bandwidth) \n*Type: boolean*"},
			{"topic.metadata.propagation.max.ms", "30000", "Apache Kafka topic creation is asynchronous and it takes some time for a new topic to propagate throughout the cluster to all brokers. If a client requests topic metadata after manual topic creation but before the topic has been fully propagated to the broker the client is requesting metadata from, the topic will seem to be non-existent and the client will mark the topic as such, failing queued produced messages with `ERR__UNKNOWN_TOPIC`. This setting delays marking a topic as non-existent until the configured propagation max time has passed. The maximum propagation time is calculated from the time the topic is first referenced in the client, e.g., on produce(). \n*Type: integer*"},
			{"topic.blacklist", "", "Topic blacklist, a comma-separated list of regular expressions for matching topic names that should be ignored in broker metadata information as if the topics did not exist. \n*Type: pattern list*"},
			{"debug", "", "A comma-separated list of debug contexts to enable. Detailed Producer debugging: broker,topic,msg. Consumer: consumer,cgrp,topic,fetch \n*Type: CSV flags*"},
			{"socket.timeout.ms", "60000", "Default timeout for network requests. Producer: ProduceRequests will use the lesser value of `socket.timeout.ms` and remaining `message.timeout.ms` for the first message in the batch. Consumer: FetchRequests will use `fetch.wait.max.ms` + `socket.timeout.ms`. Admin: Admin requests will use `socket.timeout.ms` or explicitly set `rd_kafka_AdminOptions_set_operation_timeout()` value. \n*Type: integer*"},
			{"socket.blocking.max.ms", "1000", "**DEPRECATED** No longer used. \n*Type: integer*"},
			{"socket.send.buffer.bytes", "0", "Broker socket send buffer size. System default is used if 0. \n*Type: integer*"},
			{"socket.receive.buffer.bytes", "0", "Broker socket receive buffer size. System default is used if 0. \n*Type: integer*"},
			{"socket.keepalive.enable", "false", "Enable TCP keep-alives (SO_KEEPALIVE) on broker sockets \n*Type: boolean*"},
			{"socket.nagle.disable", "false", "Disable the Nagle algorithm (TCP_NODELAY) on broker sockets. \n*Type: boolean*"},
			{"socket.max.fails", "1", "Disconnect from broker when this number of send failures (e.g., timed out requests) is reached. Disable with 0. WARNING: It is highly recommended to leave this setting at its default value of 1 to avoid the client and broker to become desynchronized in case of request timeouts. NOTE: The connection is automatically re-established. \n*Type: integer*"},
			{"broker.address.ttl", "1000", "How long to cache the broker address resolving results (milliseconds). \n*Type: integer*"},
			{"broker.address.family", "any", "Allowed broker IP address families: any, v4, v6 \n*Type: enum value*"},
			{"socket.connection.setup.timeout.ms", "30000", "Maximum time allowed for broker connection setup (TCP connection setup as well SSL and SASL handshake). If the connection to the broker is not fully functional after this the connection will be closed and retried. \n*Type: integer*"},
			{"connections.max.idle.ms", "0", "Close broker connections after the specified time of inactivity. Disable with 0. If this property is left at its default value some heuristics are performed to determine a suitable default value, this is currently limited to identifying brokers on Azure (see librdkafka issue #3109 for more info). \n*Type: integer*"},
			{"reconnect.backoff.jitter.ms", "0", "**DEPRECATED** No longer used. See `reconnect.backoff.ms` and `reconnect.backoff.max.ms`. \n*Type: integer*"},
			{"reconnect.backoff.ms", "100", "The initial time to wait before reconnecting to a broker after the connection has been closed. The time is increased exponentially until `reconnect.backoff.max.ms` is reached. -25% to +50% jitter is applied to each reconnect backoff. A value of 0 disables the backoff and reconnects immediately. \n*Type: integer*"},
			{"reconnect.backoff.max.ms", "10000", "The maximum time to wait before reconnecting to a broker after the connection has been closed. \n*Type: integer*"},
			{"statistics.interval.ms", "0", "librdkafka statistics emit interval. The application also needs to register a stats callback using `rd_kafka_conf_set_stats_cb()`. The granularity is 1000ms. A value of 0 disables statistics. \n*Type: integer*"},
			{"enabled_events", "0", "See `rd_kafka_conf_set_events()` \n*Type: integer*"},
			{"error_cb", "", "Error callback (set with rd_kafka_conf_set_error_cb()) \n*Type: see dedicated API*"},
			{"throttle_cb", "", "Throttle callback (set with rd_kafka_conf_set_throttle_cb()) \n*Type: see dedicated API*"},
			{"stats_cb", "", "Statistics callback (set with rd_kafka_conf_set_stats_cb()) \n*Type: see dedicated API*"},
			{"log_cb", "", "Log callback (set with rd_kafka_conf_set_log_cb()) \n*Type: see dedicated API*"},
			{"log_level", "6", "Logging level (syslog(3) levels) \n*Type: integer*"},
			{"log.queue", "false", "Disable spontaneous log_cb from internal librdkafka threads, instead enqueue log messages on queue set with `rd_kafka_set_log_queue()` and serve log callbacks or events through the standard poll APIs. **NOTE**: Log messages will linger in a temporary queue until the log queue has been set. \n*Type: boolean*"},
			{"log.thread.name", "true", "Print internal thread name in log messages (useful for debugging librdkafka internals) \n*Type: boolean*"},
			{"enable.random.seed", "true", "If enabled librdkafka will initialize the PRNG with srand(current_time.milliseconds) on the first invocation of rd_kafka_new() (required only if rand_r() is not available on your platform). If disabled the application must call srand() prior to calling rd_kafka_new(). \n*Type: boolean*"},
			{"log.connection.close", "true", "Log broker disconnects. It might be useful to turn this off when interacting with 0.9 brokers with an aggressive `connections.max.idle.ms` value. \n*Type: boolean*"},
			{"background_event_cb", "", "Background queue event callback (set with rd_kafka_conf_set_background_event_cb()) \n*Type: see dedicated API*"},
			{"socket_cb", "", "Socket creation callback to provide race-free CLOEXEC \n*Type: see dedicated API*"},
			{"connect_cb", "", "Socket connect callback \n*Type: see dedicated API*"},
			{"closesocket_cb", "", "Socket close callback \n*Type: see dedicated API*"},
			{"open_cb", "", "File open callback to provide race-free CLOEXEC \n*Type: see dedicated API*"},
			{"resolve_cb", "", "Address resolution callback (set with rd_kafka_conf_set_resolve_cb()). \n*Type: see dedicated API*"},
			{"opaque", "", "Application opaque (set with rd_kafka_conf_set_opaque()) \n*Type: see dedicated API*"},
			{"default_topic_conf", "", "Default topic configuration for automatically subscribed topics \n*Type: see dedicated API*"},
			{"internal.termination.signal", "0", "Signal that librdkafka will use to quickly terminate on rd_kafka_destroy(). If this signal is not set then there will be a delay before rd_kafka_wait_destroyed() returns true as internal threads are timing out their system calls. If this signal is set however the delay will be minimal. The application should mask this signal as an internal signal handler is installed. \n*Type: integer*"},
			{"api.version.request", "true", "Request broker's supported API versions to adjust functionality to available protocol features. If set to false, or the ApiVersionRequest fails, the fallback version `broker.version.fallback` will be used. **NOTE**: Depends on broker version >=0.10.0. If the request is not supported by (an older) broker the `broker.version.fallback` fallback is used. \n*Type: boolean*"},
			{"api.version.request.timeout.ms", "10000", "Timeout for broker API version requests. \n*Type: integer*"},
			{"api.version.fallback.ms", "0", "Dictates how long the `broker.version.fallback` fallback is used in the case the ApiVersionRequest fails. **NOTE**: The ApiVersionRequest is only issued when a new connection to the broker is made (such as after an upgrade). \n*Type: integer*"},
			{"broker.version.fallback", "0.10.0", "Older broker versions (before 0.10.0) provide no way for a client to query for supported protocol features (ApiVersionRequest, see `api.version.request`) making it impossible for the client to know what features it may use. As a workaround a user may set this property to the expected broker version and the client will automatically adjust its feature set accordingly if the ApiVersionRequest fails (or is disabled). The fallback broker version will be used for `api.version.fallback.ms`. Valid values are: 0.9.0, 0.8.2, 0.8.1, 0.8.0. Any other value >= 0.10, such as 0.10.2.1, enables ApiVersionRequests. \n*Type: string*"},
			{"allow.auto.create.topics", "false", "Allow automatic topic creation on the broker when subscribing to or assigning non-existent topics. The broker must also be configured with `auto.create.topics.enable=true` for this configuration to take effect. Note: the default value (true) for the producer is different from the default value (false) for the consumer. Further, the consumer default value is different from the Java consumer (true), and this property is not supported by the Java producer. Requires broker version >= 0.11.0.0, for older broker versions only the broker configuration applies. \n*Type: boolean*"},
			{"security.protocol", "plaintext", "Protocol used to communicate with brokers. \n*Type: enum value*"},
			{"ssl.cipher.suites", "", "A cipher suite is a named combination of authentication, encryption, MAC and key exchange algorithm used to negotiate the security settings for a network connection using TLS or SSL network protocol. See manual page for `ciphers(1)` and `SSL_CTX_set_cipher_list(3). \n*Type: string*"},
			{"ssl.curves.list", "", "The supported-curves extension in the TLS ClientHello message specifies the curves (standard/named, or 'explicit' GF(2^k) or GF(p)) the client is willing to have the server use. See manual page for `SSL_CTX_set1_curves_list(3)`. OpenSSL >= 1.0.2 required. \n*Type: string*"},
			{"ssl.sigalgs.list", "", "The client uses the TLS ClientHello signature_algorithms extension to indicate to the server which signature/hash algorithm pairs may be used in digital signatures. See manual page for `SSL_CTX_set1_sigalgs_list(3)`. OpenSSL >= 1.0.2 required. \n*Type: string*"},
			{"ssl.key.location", "", "Path to client's private key (PEM) used for authentication. \n*Type: string*"},
			{"ssl.key.password", "", "Private key passphrase (for use with `ssl.key.location` and `set_ssl_cert()`) \n*Type: string*"},
			{"ssl.key.pem", "", "Client's private key string (PEM format) used for authentication. \n*Type: string*"},
			{"ssl_key", "", "Client's private key as set by rd_kafka_conf_set_ssl_cert() \n*Type: see dedicated API*"},
			{"ssl.certificate.location", "", "Path to client's public key (PEM) used for authentication. \n*Type: string*"},
			{"ssl.certificate.pem", "", "Client's public key string (PEM format) used for authentication. \n*Type: string*"},
			{"ssl_certificate", "", "Client's public key as set by rd_kafka_conf_set_ssl_cert() \n*Type: see dedicated API*"},
			{"ssl.ca.location", "", "File or directory path to CA certificate(s) for verifying the broker's key. Defaults: On Windows the system's CA certificates are automatically looked up in the Windows Root certificate store. On Mac OSX this configuration defaults to `probe`. It is recommended to install openssl using Homebrew, to provide CA certificates. On Linux install the distribution's ca-certificates package. If OpenSSL is statically linked or `ssl.ca.location` is set to `probe` a list of standard paths will be probed and the first one found will be used as the default CA certificate location path. If OpenSSL is dynamically linked the OpenSSL library's default path will be used (see `OPENSSLDIR` in `openssl version -a`). \n*Type: string*"},
			{"ssl.ca.pem", "", "CA certificate string (PEM format) for verifying the broker's key. \n*Type: string*"},
			{"ssl_ca", "", "CA certificate as set by rd_kafka_conf_set_ssl_cert() \n*Type: see dedicated API*"},
			{"ssl.ca.certificate.stores", "Root", "Comma-separated list of Windows Certificate stores to load CA certificates from. Certificates will be loaded in the same order as stores are specified. If no certificates can be loaded from any of the specified stores an error is logged and the OpenSSL library's default CA location is used instead. Store names are typically one or more of: MY, Root, Trust, CA. \n*Type: string*"},
			{"ssl.crl.location", "", "Path to CRL for verifying broker's certificate validity. \n*Type: string*"},
			{"ssl.keystore.location", "", "Path to client's keystore (PKCS#12) used for authentication. \n*Type: string*"},
			{"ssl.keystore.password", "", "Client's keystore (PKCS#12) password. \n*Type: string*"},
			{"ssl.providers", "", "Comma-separated list of OpenSSL 3.0.x implementation providers. E.g., default,legacy. \n*Type: string*"},
			{"ssl.engine.location", "", "**DEPRECATED** Path to OpenSSL engine library. OpenSSL >= 1.1.x required. DEPRECATED: OpenSSL engine support is deprecated and should be replaced by OpenSSL 3 providers. \n*Type: string*"},
			{"ssl.engine.id", "dynamic", "OpenSSL engine id is the name used for loading engine. \n*Type: string*"},
			{"ssl_engine_callback_data", "", "OpenSSL engine callback data (set with rd_kafka_conf_set_engine_callback_data()). \n*Type: see dedicated API*"},
			{"enable.ssl.certificate.verification", "true", "Enable OpenSSL's builtin broker (server) certificate verification. This verification can be extended by the application by implementing a certificate_verify_cb. \n*Type: boolean*"},
			{"ssl.endpoint.identification.algorithm", "https", "Endpoint identification algorithm to validate broker hostname using broker certificate. https - Server (broker) hostname verification as specified in RFC2818. none - No endpoint verification. OpenSSL >= 1.0.2 required. \n*Type: enum value*"},
			{"ssl.certificate.verify_cb", "", "Callback to verify the broker certificate chain. \n*Type: see dedicated API*"},
			{"sasl.mechanisms", "GSSAPI", "SASL mechanism to use for authentication. Supported: GSSAPI, PLAIN, SCRAM-SHA-256, SCRAM-SHA-512, OAUTHBEARER. **NOTE**: Despite the name only one mechanism must be configured. \n*Type: string*"},
			{"sasl.mechanism", "GSSAPI", "Alias for `sasl.mechanisms`: SASL mechanism to use for authentication. Supported: GSSAPI, PLAIN, SCRAM-SHA-256, SCRAM-SHA-512, OAUTHBEARER. **NOTE**: Despite the name only one mechanism must be configured. \n*Type: string*"},
			{"sasl.kerberos.service.name", "kafka", "Kerberos principal name that Kafka runs as, not including /hostname@REALM \n*Type: string*"},
			{"sasl.kerberos.principal", "kafkaclient", "This client's Kerberos principal name. (Not supported on Windows, will use the logon user's principal). \n*Type: string*"},
			{"sasl.kerberos.kinit.cmd", "kinit -R -t \"%{sasl.kerberos.keytab}\" -k %{sasl.kerberos.principal} || kinit -t \"%{sasl.kerberos.keytab}\" -k %{sasl.kerberos.principal}", "Shell command to refresh or acquire the client's Kerberos ticket. This command is executed on client creation and every sasl.kerberos.min.time.before.relogin (0=disable). %{config.prop.name} is replaced by corresponding config object value. \n*Type: string*"},
			{"sasl.kerberos.keytab", "", "Path to Kerberos keytab file. This configuration property is only used as a variable in `sasl.kerberos.kinit.cmd` as ` ... -t \"%{sasl.kerberos.keytab}\"`. \n*Type: string*"},
			{"sasl.kerberos.min.time.before.relogin", "60000", "Minimum time in milliseconds between key refresh attempts. Disable automatic key refresh by setting this property to 0. \n*Type: integer*"},
			{"sasl.username", "", "SASL username for use with the PLAIN and SASL-SCRAM-.. mechanisms \n*Type: string*"},
			{"sasl.password", "", "SASL password for use with the PLAIN and SASL-SCRAM-.. mechanism \n*Type: string*"},
			{"sasl.oauthbearer.config", "", "SASL/OAUTHBEARER configuration. The format is implementation-dependent and must be parsed accordingly. The default unsecured token implementation (see https://tools.ietf.org/html/rfc7515#appendix-A.5) recognizes space-separated name=value pairs with valid names including principalClaimName, principal, scopeClaimName, scope, and lifeSeconds. The default value for principalClaimName is \"sub\", the default value for scopeClaimName is \"scope\", and the default value for lifeSeconds is 3600. The scope value is CSV format with the default value being no/empty scope. For example: `principalClaimName=azp principal=admin scopeClaimName=roles scope=role1,role2 lifeSeconds=600`. In addition, SASL extensions can be communicated to the broker via `extension_NAME=value`. For example: `principal=admin extension_traceId=123` \n*Type: string*"},
			{"enable.sasl.oauthbearer.unsecure.jwt", "false", "Enable the builtin unsecure JWT OAUTHBEARER token handler if no oauthbearer_refresh_cb has been set. This builtin handler should only be used for development or testing, and not in production. \n*Type: boolean*"},
			{"oauthbearer_token_refresh_cb", "", "SASL/OAUTHBEARER token refresh callback (set with rd_kafka_conf_set_oauthbearer_token_refresh_cb(), triggered by rd_kafka_poll(), et.al. This callback will be triggered when it is time to refresh the client's OAUTHBEARER token. Also see `rd_kafka_conf_enable_sasl_queue()`. \n*Type: see dedicated API*"},
			{"sasl.oauthbearer.method", "default", "Set to \"default\" or \"oidc\" to control which login method to be used. If set to \"oidc\", the following properties must also be be specified: `sasl.oauthbearer.client.id`, `sasl.oauthbearer.client.secret`, and `sasl.oauthbearer.token.endpoint.url`. \n*Type: enum value*"},
			{"sasl.oauthbearer.client.id", "", "Public identifier for the application. Must be unique across all clients that the authorization server handles. Only used when `sasl.oauthbearer.method` is set to \"oidc\". \n*Type: string*"},
			{"sasl.oauthbearer.client.secret", "", "Client secret only known to the application and the authorization server. This should be a sufficiently random string that is not guessable. Only used when `sasl.oauthbearer.method` is set to \"oidc\". \n*Type: string*"},
			{"sasl.oauthbearer.scope", "", "Client use this to specify the scope of the access request to the broker. Only used when `sasl.oauthbearer.method` is set to \"oidc\". \n*Type: string*"},
			{"sasl.oauthbearer.extensions", "", "Allow additional information to be provided to the broker. Comma-separated list of key=value pairs. E.g., \"supportFeatureX=true,organizationId=sales-emea\".Only used when `sasl.oauthbearer.method` is set to \"oidc\". \n*Type: string*"},
			{"sasl.oauthbearer.token.endpoint.url", "", "OAuth/OIDC issuer token endpoint HTTP(S) URI used to retrieve token. Only used when `sasl.oauthbearer.method` is set to \"oidc\". \n*Type: string*"},
			{"plugin.library.paths", "", "List of plugin libraries to load (; separated). The library search path is platform dependent (see dlopen(3) for Unix and LoadLibrary() for Windows). If no filename extension is specified the platform-specific extension (such as .dll or .so) will be appended automatically. \n*Type: string*"},
			{"interceptors", "", "Interceptors added through rd_kafka_conf_interceptor_add_..() and any configuration handled by interceptors. \n*Type: see dedicated API*"},
			{"client.dns.lookup", "use_all_dns_ips", "Controls how the client uses DNS lookups. By default, when the lookup returns multiple IP addresses for a hostname, they will all be attempted for connection before the connection is considered failed. This applies to both bootstrap and advertised servers. If the value is set to `resolve_canonical_bootstrap_servers_only`, each entry will be resolved and expanded into a list of canonical names. **WARNING**: `resolve_canonical_bootstrap_servers_only` must only be used with `GSSAPI` (Kerberos) as `sasl.mechanism`, as it's the only purpose of this configuration value. **NOTE**: Default here is different from the Java client's default behavior, which connects only to the first IP address returned for a hostname.  \n*Type: enum value*"},
			{"enable.metrics.push", "true", "Whether to enable pushing of client metrics to the cluster, if the cluster has a client metrics subscription which matches this client \n*Type: boolean*"},
			{"client.rack", "", "A rack identifier for this client. This can be any string value which indicates where this client is physically located. It corresponds with the broker config `broker.rack`. \n*Type: string*"},
			{"retry.backoff.ms", "100", "The backoff time in milliseconds before retrying a protocol request, this is the first backoff time, and will be backed off exponentially until number of retries is exhausted, and it's capped by retry.backoff.max.ms. \n*Type: integer*"},
			{"retry.backoff.max.ms", "1000", "The max backoff time in milliseconds before retrying a protocol request, this is the atmost backoff allowed for exponentially backed off requests. \n*Type: integer*"},
			{"opaque", "", "Application opaque (set with rd_kafka_topic_conf_set_opaque()) \n*Type: see dedicated API*"},
			{"group.id", "", "Client group id string. All clients sharing the same group.id belong to the same group. \n*Type: string*"},
			{"group.instance.id", "", "Enable static group membership. Static group members are able to leave and rejoin a group within the configured `session.timeout.ms` without prompting a group rebalance. This should be used in combination with a larger `session.timeout.ms` to avoid group rebalances caused by transient unavailability (e.g. process restarts). Requires broker version >= 2.3.0. \n*Type: string*"},
			{"partition.assignment.strategy", "range,roundrobin", "The name of one or more partition assignment strategies. The elected group leader will use a strategy supported by all members of the group to assign partitions to group members. If there is more than one eligible strategy, preference is determined by the order of this list (strategies earlier in the list have higher priority). Cooperative and non-cooperative (eager) strategies must not be mixed. Available strategies: range, roundrobin, cooperative-sticky. \n*Type: string*"},
			{"session.timeout.ms", "45000", "Client group session and failure detection timeout. The consumer sends periodic heartbeats (heartbeat.interval.ms) to indicate its liveness to the broker. If no hearts are received by the broker for a group member within the session timeout, the broker will remove the consumer from the group and trigger a rebalance. The allowed range is configured with the **broker** configuration properties `group.min.session.timeout.ms` and `group.max.session.timeout.ms`. Also see `max.poll.interval.ms`. \n*Type: integer*"},
			{"heartbeat.interval.ms", "3000", "Group session keepalive heartbeat interval. \n*Type: integer*"},
			{"group.protocol.type", "consumer", "Group protocol type for the `classic` group protocol. NOTE: Currently, the only supported group protocol type is `consumer`. \n*Type: string*"},
			{"group.protocol", "classic", "Group protocol to use. Use `classic` for the original protocol and `consumer` for the new protocol introduced in KIP-848. Available protocols: classic or consumer. Default is `classic`, but will change to `consumer` in next releases. \n*Type: enum value*"},
			{"group.remote.assignor", "", "Server side assignor to use. Keep it null to make server select a suitable assignor for the group. Available assignors: uniform or range. Default is null \n*Type: string*"},
			{"coordinator.query.interval.ms", "600000", "How often to query for the current client group coordinator. If the currently assigned coordinator is down the configured query interval will be divided by ten to more quickly recover in case of coordinator reassignment. \n*Type: integer*"},
			{"max.poll.interval.ms", "300000", "Maximum allowed time between calls to consume messages (e.g., rd_kafka_consumer_poll()) for high-level consumers. If this interval is exceeded the consumer is considered failed and the group will rebalance in order to reassign the partitions to another consumer group member. Warning: Offset commits may be not possible at this point. Note: It is recommended to set `enable.auto.offset.store=false` for long-time processing applications and then explicitly store offsets (using offsets_store()) *after* message processing, to make sure offsets are not auto-committed prior to processing has finished. The interval is checked two times per second. See KIP-62 for more information. \n*Type: integer*"},
			{"enable.auto.commit", "true", "Automatically and periodically commit offsets in the background. Note: setting this to false does not prevent the consumer from fetching previously committed start offsets. To circumvent this behaviour set specific start offsets per partition in the call to assign(). \n*Type: boolean*"},
			{"auto.commit.interval.ms", "5000", "The frequency in milliseconds that the consumer offsets are committed (written) to offset storage. (0 = disable). This setting is used by the high-level consumer. \n*Type: integer*"},
			{"enable.auto.offset.store", "true", "Automatically store offset of last message provided to application. The offset store is an in-memory store of the next offset to (auto-)commit for each partition. \n*Type: boolean*"},
			{"queued.min.messages", "100000", "Minimum number of messages per topic+partition librdkafka tries to maintain in the local consumer queue. \n*Type: integer*"},
			{"queued.max.messages.kbytes", "65536", "Maximum number of kilobytes of queued pre-fetched messages in the local consumer queue. If using the high-level consumer this setting applies to the single consumer queue, regardless of the number of partitions. When using the legacy simple consumer or when separate partition queues are used this setting applies per partition. This value may be overshot by fetch.message.max.bytes. This property has higher priority than queued.min.messages. \n*Type: integer*"},
			{"fetch.wait.max.ms", "500", "Maximum time the broker may wait to fill the Fetch response with fetch.min.bytes of messages. \n*Type: integer*"},
			{"fetch.queue.backoff.ms", "1000", "How long to postpone the next fetch request for a topic+partition in case the current fetch queue thresholds (queued.min.messages or queued.max.messages.kbytes) have been exceded. This property may need to be decreased if the queue thresholds are set low and the application is experiencing long (~1s) delays between messages. Low values may increase CPU utilization. \n*Type: integer*"},
			{"fetch.message.max.bytes", "1048576", "Initial maximum number of bytes per topic+partition to request when fetching messages from the broker. If the client encounters a message larger than this value it will gradually try to increase it until the entire message can be fetched. \n*Type: integer*"},
			{"max.partition.fetch.bytes", "1048576", "Alias for `fetch.message.max.bytes`: Initial maximum number of bytes per topic+partition to request when fetching messages from the broker. If the client encounters a message larger than this value it will gradually try to increase it until the entire message can be fetched. \n*Type: integer*"},
			{"fetch.max.bytes", "52428800", "Maximum amount of data the broker shall return for a Fetch request. Messages are fetched in batches by the consumer and if the first message batch in the first non-empty partition of the Fetch request is larger than this value, then the message batch will still be returned to ensure the consumer can make progress. The maximum message batch size accepted by the broker is defined via `message.max.bytes` (broker config) or `max.message.bytes` (broker topic config). `fetch.max.bytes` is automatically adjusted upwards to be at least `message.max.bytes` (consumer config). \n*Type: integer*"},
			{"fetch.min.bytes", "1", "Minimum number of bytes the broker responds with. If fetch.wait.max.ms expires the accumulated data will be sent to the client regardless of this setting. \n*Type: integer*"},
			{"fetch.error.backoff.ms", "500", "How long to postpone the next fetch request for a topic+partition in case of a fetch error. \n*Type: integer*"},
			{"offset.store.method", "broker", "**DEPRECATED** Offset commit store method: 'file' - DEPRECATED: local file store (offset.store.path, et.al), 'broker' - broker commit store (requires Apache Kafka 0.8.2 or later on the broker). \n*Type: enum value*"},
			{"isolation.level", "read_committed", "Controls how to read messages written transactionally: `read_committed` - only return transactional messages which have been committed. `read_uncommitted` - return all messages, even transactional messages which have been aborted. \n*Type: enum value*"},
			{"consume_cb", "", "Message consume callback (set with rd_kafka_conf_set_consume_cb()) \n*Type: see dedicated API*"},
			{"rebalance_cb", "", "Called after consumer group has been rebalanced (set with rd_kafka_conf_set_rebalance_cb()) \n*Type: see dedicated API*"},
			{"offset_commit_cb", "", "Offset commit result propagation callback. (set with rd_kafka_conf_set_offset_commit_cb()) \n*Type: see dedicated API*"},
			{"enable.partition.eof", "false", "Emit RD_KAFKA_RESP_ERR__PARTITION_EOF event whenever the consumer reaches the end of a partition. \n*Type: boolean*"},
			{"check.crcs", "false", "Verify CRC32 of consumed messages, ensuring no on-the-wire or on-disk corruption to the messages occurred. This check comes at slightly increased CPU usage. \n*Type: boolean*"},
			{"auto.commit.enable", "true", "**DEPRECATED** [**LEGACY PROPERTY:** This property is used by the simple legacy consumer only. When using the high-level KafkaConsumer, the global `enable.auto.commit` property must be used instead]. If true, periodically commit offset of the last message handed to the application. This committed offset will be used when the process restarts to pick up where it left off. If false, the application will have to call `rd_kafka_offset_store()` to store an offset (optional). Offsets will be written to broker or local file according to offset.store.method. \n*Type: boolean*"},
			{"enable.auto.commit", "true", "**DEPRECATED** Alias for `auto.commit.enable`: [**LEGACY PROPERTY:** This property is used by the simple legacy consumer only. When using the high-level KafkaConsumer, the global `enable.auto.commit` property must be used instead]. If true, periodically commit offset of the last message handed to the application. This committed offset will be used when the process restarts to pick up where it left off. If false, the application will have to call `rd_kafka_offset_store()` to store an offset (optional). Offsets will be written to broker or local file according to offset.store.method. \n*Type: boolean*"},
			{"auto.commit.interval.ms", "60000", "[**LEGACY PROPERTY:** This setting is used by the simple legacy consumer only. When using the high-level KafkaConsumer, the global `auto.commit.interval.ms` property must be used instead]. The frequency in milliseconds that the consumer offsets are committed (written) to offset storage. \n*Type: integer*"},
			{"auto.offset.reset", "largest", "Action to take when there is no initial offset in offset store or the desired offset is out of range: 'smallest','earliest' - automatically reset the offset to the smallest offset, 'largest','latest' - automatically reset the offset to the largest offset, 'error' - trigger an error (ERR__AUTO_OFFSET_RESET) which is retrieved by consuming messages and checking 'message->err'. \n*Type: enum value*"},
			{"offset.store.path", ".", "**DEPRECATED** Path to local file for storing offsets. If the path is a directory a filename will be automatically generated in that directory based on the topic and partition. File-based offset storage will be removed in a future version. \n*Type: string*"},
			{"offset.store.sync.interval.ms", "-1", "**DEPRECATED** fsync() interval for the offset file, in milliseconds. Use -1 to disable syncing, and 0 for immediate sync after each write. File-based offset storage will be removed in a future version. \n*Type: integer*"},
			{"offset.store.method", "broker", "**DEPRECATED** Offset commit store method: 'file' - DEPRECATED: local file store (offset.store.path, et.al), 'broker' - broker commit store (requires \"group.id\" to be configured and Apache Kafka 0.8.2 or later on the broker.). \n*Type: enum value*"},
			{"consume.callback.max.messages", "0", "Maximum number of messages to dispatch in one `rd_kafka_consume_callback*()` call (0 = unlimited) \n*Type: integer*"},
		},
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet, prefix string) {
	if prefix != "" {
		prefix = prefix + "."
	}

	for _, o := range o.Opts {
		fs.StringVar(&o.Value, prefix+strings.ReplaceAll(o.Name, ".", "-"), o.Value, o.Description)
	}
}

func (o *Options) Validate() []error {
	var errs []error

	return errs
}

func (o *Options) Complete() []error {
	var errs []error

	return errs
}
