package secretshare

import "errors"

/**
 * Abstract class for secret sharing schemes.
 *
 * <p>
 * The abstract class <code>SecretSharingScheme</code> provides default methods including setting access structure,
 * generating shares from secret and calculating secret from shares. And also an initialization checking method is
 * provided since it should be checked before generating shares and calculating secret. Subclasses of <code>SecretSharingScheme</code>
 * should override the above methods.
 * <p>
 * Echo participant in the secret sharing system has an unique int ID, starting from 0 to (N - 1) (N is the
 * number of participants). The share generating method outputs N shares, one for each participant which contains the ID of
 * the participant and its shared value.
 *
 * @author 		LoCCS
 * @version		1.0
 *
 */

type SecretSharingScheme struct {
	/**
	 * The number of participants that share the secret.
	 */
	participantCount int

	/**
	 * Access structure
	 */
	access AccessStructureInterface

	/**
   * Abstract Interfaces of SecretSharingScheme
   */
	SecretSharingSchemeITF SecretSharingSchemeInterface
}

type SecretSharingSchemeInterface interface {

	GetParticipantCount() int

	GetAccessStructure() AccessStructureInterface

	GenerateShares(secret interface{}, auxiliary []interface{}) ([]*SecretShare, error)

	CalculateSecret(shares []*SecretShare) (interface{}, error)

	/**
    * Set access structure.
    * <p>
    * The participant count in AccessStructure and that in this scheme must be equal.
    *
    * @param access Access structure.
    * @return error If the access structure is invalid.
    */
	SetAccessStructure(access AccessStructureInterface) error

	/**
    * Determine if the scheme object is initialized properly for generating shares and calculating secret.
    *
    * @return True if the scheme object is initialized properly, otherwise return false.
    */
	IsInitialized() bool

	/**
 	* Abstract method of generating shares from input secret.
 	*
 	* @param secret The secret from which shares are generated.
 	* @param auxiliary Auxiliary data for generating shares.
 	* @return N shares that generated from the input secret, one for each participant.
	* @return error If the input secret is invalid.
 	*/
 	generateSharesImpl(secret interface{}, auxiliary []interface{}) ([]*SecretShare, error)

	/**
 	* Abstract method of calculating secret from input shares.
 	*
 	* @param shares The shares from which secret is calculated.
 	* @return The secret calculated from the input shares.
 	* @return error If any of the input shares is invalid.
 	*/
 	calculateSecretImpl(shares []*SecretShare) (interface{},error)
}

/**
 * Get participant count.
 *
 * @return participantCount, type int.
 */
func (sss *SecretSharingScheme) GetParticipantCount() int{
	return sss.participantCount
}

/**
 * Get access structure.
 *
 * @return AccessStructureInterface.
 */
func (sss *SecretSharingScheme) GetAccessStructure() AccessStructureInterface{
	return sss.access
}

/**
 * Generate shares from the input secret.
 * <p>
 * Call implemented <code>generateSharesImpl</code> method to generate shares.
 *
 * @param secret The secret from which shares are generated.
 * @param auxiliary Auxiliary data for generating shares.
 * @return N shares that generated from the input secret, one for each participant.
 * @return error error if the input secret is invalid or the scheme object is not ready.
 */
func (sss *SecretSharingScheme) GenerateShares(secret interface{}, auxiliary []interface{}) ([]*SecretShare, error){
	if (!sss.SecretSharingSchemeITF.IsInitialized()) {
		return nil, errors.New("Not ready for generate shares.")
	}
	return sss.SecretSharingSchemeITF.generateSharesImpl(secret,auxiliary)
}

/**
 * Abstract method of calculating secret from input shares.
 *
 * @param shares The shares from which secret is calculated.
 * @return The secret calculated from the input shares.
 * @return error If any of the input shares is invalid.
 */
func (sss *SecretSharingScheme) CalculateSecret(shares []*SecretShare) (interface{}, error){
	if (!sss.SecretSharingSchemeITF.IsInitialized()){
		return nil, errors.New("Not ready for calculating secret.")
	}
	participants := make([]int, len(shares))
	for i := 0; i < len(shares); i++{
		participants[i] = shares[i].GetParticipant()
	}
	ok, err := sss.access.TestAccessable(participants)
	if (err != nil) {return nil, err}
	if (!ok) {return nil, errors.New("Access Test not passed.")}
	return sss.SecretSharingSchemeITF.calculateSecretImpl(shares)
}
