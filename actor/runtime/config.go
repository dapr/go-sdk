package runtime

type ActorRuntimeConfig struct {
	RegisteredActorTypes []string `json:"entities"`
	ActorIdleTimeout string `json:"actorIdleTimeout"`
	ActorScanInterval string `json:"actorScanInterval"`
	DrainOngingCallTimeout string `json:"drainOngoingCallTimeout"`
	DrainBalancedActors bool `json:"drainRebalancedActors"`
}

