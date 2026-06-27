package turn_detector

// TurnDetectorModel names a turn-detection model. It is an alias for string,
// so callers may pass a literal ("namo") or one of the constants below.
type TurnDetectorModel = string

// Models selectable via TurnDetectorOptions.Model.
const (
	// ModelNamo is the default Namo detector. It is the only model that honors
	// TurnDetectorOptions.Language.
	ModelNamo TurnDetectorModel = "namo"
	// ModelNamoInference is the hosted Namo detector.
	ModelNamoInference TurnDetectorModel = "namo-inference"
	// ModelEchoSmall is the small Echo detector.
	ModelEchoSmall TurnDetectorModel = "echo-small"
	// ModelEchoLarge is the large Echo detector.
	ModelEchoLarge TurnDetectorModel = "echo-large"
)
