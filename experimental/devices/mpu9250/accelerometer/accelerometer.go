package accelerometer

const (
	ACCEL_FS_SEL_2G  = 0
	ACCEL_FS_SEL_4G  = 8
	ACCEL_FS_SEL_8G  = 0x10
	ACCEL_FS_SEL_16G = 0x18

	ACCEL_FS_SENS_2G  = 2.0 / 32768.0
	ACCEL_FS_SENS_4G  = 4.0 / 32768.0
	ACCEL_FS_SENS_8G  = 8.0 / 32768.0
	ACCEL_FS_SENS_16G = 16.0 / 32768.0
)

func Sensitivity(selector int) float32 {
	switch selector {
	case ACCEL_FS_SEL_2G:
		return ACCEL_FS_SENS_2G
	case ACCEL_FS_SEL_4G:
		return ACCEL_FS_SENS_4G
	case ACCEL_FS_SEL_8G:
		return ACCEL_FS_SENS_8G
	case ACCEL_FS_SEL_16G:
		return ACCEL_FS_SENS_16G
	}
	return ACCEL_FS_SENS_2G
}
