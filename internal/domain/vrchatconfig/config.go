package vrchatconfig

// VRChatConfig represents the VRChat config.json structure.
// Reference: https://docs.vrchat.com/docs/configuration-file
type VRChatConfig struct {
	// Camera & Screenshot
	CameraResWidth           int    `json:"camera_res_width,omitempty"`
	CameraResHeight          int    `json:"camera_res_height,omitempty"`
	ScreenshotResWidth       int    `json:"screenshot_res_width,omitempty"`
	ScreenshotResHeight      int    `json:"screenshot_res_height,omitempty"`
	PictureOutputFolder      string `json:"picture_output_folder,omitempty"`
	PictureOutputSplitByDate *bool  `json:"picture_output_split_by_date,omitempty"`
	FPVSteadycamFOV          int    `json:"fpv_steadycam_fov,omitempty"`

	// Cache
	CacheDirectory   string `json:"cache_directory,omitempty"`
	CacheSize        int    `json:"cache_size,omitempty"`
	CacheExpiryDelay int    `json:"cache_expiry_delay,omitempty"`

	// Rich Presence
	DisableRichPresence *bool `json:"disableRichPresence,omitempty"`

	// Particle System Limits (betas required)
	Betas                   []string `json:"betas,omitempty"`
	PSMaxParticles          int      `json:"ps_max_particles,omitempty"`
	PSMaxSystems            int      `json:"ps_max_systems,omitempty"`
	PSMaxEmission           int      `json:"ps_max_emission,omitempty"`
	PSMaxTotalEmission      int      `json:"ps_max_total_emission,omitempty"`
	PSMeshParticleDivider   int      `json:"ps_mesh_particle_divider,omitempty"`
	PSMeshParticlePolyLimit int      `json:"ps_mesh_particle_poly_limit,omitempty"`
	PSCollisionPenaltyHigh  int      `json:"ps_collision_penalty_high,omitempty"`
	PSCollisionPenaltyMed   int      `json:"ps_collision_penalty_med,omitempty"`
	PSCollisionPenaltyLow   int      `json:"ps_collision_penalty_low,omitempty"`
	PSTrailsPenalty         int      `json:"ps_trails_penalty,omitempty"`

	// Dynamic Bone Limits (legacy)
	DynamicBoneMaxAffectedTransformCount int `json:"dynamic_bone_max_affected_transform_count,omitempty"`
	DynamicBoneMaxColliderCheckCount     int `json:"dynamic_bone_max_collider_check_count,omitempty"`
}

// ConfigRepository provides CRUD operations for VRChat config.json.
type ConfigRepository interface {
	Exists() (bool, error)
	Read() (*VRChatConfig, error)
	Write(cfg *VRChatConfig) error
	Delete() error
}
