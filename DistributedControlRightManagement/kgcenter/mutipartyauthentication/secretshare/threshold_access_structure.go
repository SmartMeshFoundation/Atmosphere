package secretshare

import "errors"

/**
 * The class implements a threshold access structure for secret sharing.
 *
 * <p>
 * A (k, n) threshold access structure (k is the threshold, n is the number of participants) achieve the following goals.
 * <ol>
 *    <li> The secret is divided into n pieces;
 *    <li> Any k or more pieces match the access structure.
 *    <li> Any k - 1 or fewer pieces do not match the access structure.
 * </ol>
 *
 * @author 		LoCCS
 * @version		1.0
 */
type ThresholdAccessStructure struct {

	/**
    * threshold k, type int .
    */
	threshold int

	AccessStructure
}

/**
 * Construct ThresholdAccessStructure with the number of participants.
 *
 * @param participantCount The number of participants that share the secret.
 * @param threshold The threshold k.
 * @return newThresholdAccessStructure the new constructed ThresholdAccessStructure
 * @return error If the number of participants or the threshold is invalid.
 */
func NewThresholdAccessStructure(participantCount int, threshold int) (*ThresholdAccessStructure, error){
	if (participantCount < 2){
		return nil, errors.New("Invalid participant count. Should be greater than 1.")
	}
	if (threshold < 1 || threshold > participantCount){
		return nil, errors.New("Invalid threshold count. Should be larger than 0 and no more than participant count.")
	}
	newThresholdAccessStructure := new(ThresholdAccessStructure)
	newThresholdAccessStructure.participantCount = participantCount;
	newThresholdAccessStructure.threshold = threshold
	newThresholdAccessStructure.AccessStructureITF = newThresholdAccessStructure
	return newThresholdAccessStructure, nil
}

/**
 * Get threshold.
 *
 * @return threshold k.
 */
func (taccs *ThresholdAccessStructure) GetThreshold() int{
	return taccs.threshold
}

/**
 * Test if the number of participants trying to calculate secret reaches threshold.
 */
func (taccs *ThresholdAccessStructure) testAccessableImpl(participants []int) (bool){
	return len(participants) >= taccs.threshold
}
