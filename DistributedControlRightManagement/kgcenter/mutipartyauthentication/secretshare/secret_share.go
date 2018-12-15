package secretshare
/**
 * A share in secret sharing schemes.
 * <p>
 * A secret share contains the ID of the participant and its shared value.
 *
 * @author 		LoCCS
 * @version		1.0
 */

type SecretShare struct {
	/**
    * ID of the participant.
    */
	participant int

	/**
	 * The shared value for this participant.
	 */
	 value interface{}
}

/**
 * Construct SecretShare from participant ID and its shared value.
 *
 * @param participant ID of the participant.
 * @param value The shared value for this participant.
 * @return newSecretShare the new constructed SecretShare
 */
func NewSecretShare (participant int , value interface{}) (*SecretShare){
	newSecretShare := new(SecretShare)
	newSecretShare.value = value
	newSecretShare.participant = participant
	return newSecretShare
}

/**
 * Get participant ID.
 *
 * @return Participant ID.
 */
func (ss *SecretShare) GetParticipant() int{
	return ss.participant
}

/**
 * Get shared value.
 *
 * @return Shared value.
 */
func (ss *SecretShare) GetValue() (interface{}){
	return ss.value
}
