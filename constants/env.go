package constants

const (
	// AppNameEnvVar is the environment variable name of the "appName" as supplied to helm
	// if not set the default will always be AppNameDefault.
	AppNameEnvVar = "APP_NAME"

	// ManagerLoggerLevelEnv is the environment variable name that can be used to set the
	// clabernetes manager logger level. This is the logger for the "main" process, not the
	// individual controllers.
	ManagerLoggerLevelEnv = "MANAGER_LOGGER_LEVEL"

	// ControllerLoggerLevelEnv is the environment variable name that can be used to set the
	// clabernetes controllers logger level.
	ControllerLoggerLevelEnv = "CONTROLLER_LOGGER_LEVEL"

	// ClientOperationTimeoutMultiplierEnv is the multiplier applied to the default client
	// operation timeout.
	ClientOperationTimeoutMultiplierEnv = "CLIENT_OPERATION_TIMEOUT_MULTIPLIER"
)

const (
	// LauncherLoggerLevelEnv is the environment variable name that can be used to set the
	// clabernetes launcher logger level.
	LauncherLoggerLevelEnv = "LAUNCHER_LOGGER_LEVEL"

	// LauncherContainerlabDebug is the environment variable name that can be used to enable the
	// debug flag of clabernetes when invoked on the launcher pod.
	LauncherContainerlabDebug = "LAUNCHER_CONTAINERLAB_DEBUG"

	// LauncherImageEnv env var that tells the controllers what image to use for clabernetes
	// (launcher) pods.
	LauncherImageEnv = "LAUNCHER_IMAGE"

	// LauncherPullPolicyEnv env var that tells the controllers what pull policy to use for
	// clabernetes (launcher) pods.
	LauncherPullPolicyEnv = "LAUNCHER_PULL_POLICY"

	// LauncherInsecureRegistries env var that tells the launcher pods which registries are
	// insecure. Should be set by the controller via the topology spec.
	LauncherInsecureRegistries = "LAUNCHER_INSECURE_REGISTRIES"
)

const (
	// ClickerLoggerLevelEnv is the environment variable name that can be used to set the
	// cl(abernetes t)ick(l)er logger level.
	ClickerLoggerLevelEnv = "CLICKER_LOGGER_LEVEL"

	// ClickerWorkerImage is the environment variable name that can be used to set the
	// cl(abernetes t)ick(l)er worker image -- that is, the image that is deployed in a pod on all
	// target nodes, by default this is simply 'busybox'.
	ClickerWorkerImage = "CLICKER_WORKER_IMAGE"

	// ClickerWorkerCommand is the command for the worker -- defaults to "/bin/sh".
	ClickerWorkerCommand = "CLICKER_WORKER_COMMAND"

	// ClickerWorkerScript is the script for the clicker worker -- defaults to 'echo "hello, there"'
	// since we can't know what users will need here.
	ClickerWorkerScript = "CLICKER_WORKER_SCRIPT"
)
