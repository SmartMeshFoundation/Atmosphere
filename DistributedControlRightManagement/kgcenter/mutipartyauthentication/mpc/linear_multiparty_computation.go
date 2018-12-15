package mpc

import (
	"github.com/BGW-SecureLinearMultiPartyComputation/src/loccs.sjtu.edu.cn/adcrypto/secretshare"
	"errors"
)

/**
 * Abstract class for secure multi-party linear function computation.
 * <p>
 * The abstract class <code>LinearMultipartyComputation</code> generalize the linear mpc
 * scheme in "Ben-Or M, Goldwasser S, Wigderson A. Completeness theorems for non-cryptographic
 * fault-tolerant distributed computation. InProceedings of the twentieth annual ACM symposium
 * on Theory of computing 1988 Jan 1 (pp. 1-10). ACM."
 * <p>
 * In MPC, a given number of participants <i>p</i><sub>1</sub>, <i>p</i><sub>2</sub>, ..., <i>p<sub>n</sub></i>,
 * each has private data, respectively <i>x</i><sub>1</sub>, <i>x</i><sub>2</sub>, ..., <i>x<sub>n</sub></i>.
 * Participants want to compute the value of a public function on the private data:
 * <i>f</i>(<i>x</i><sub>1</sub>, <i>x</i><sub>2</sub>, ..., <i>x<sub>n</sub></i>)
 * while keeping their own inputs secret， if there are no more than <i>t</i>&lt;<i>n</i>/2 semi-honest
 * adversaries.
 * <p>
 * A linear function is in the form <i>f</i>(<i>x</i><sub>1</sub>, <i>x</i><sub>2</sub>, ..., <i>x<sub>n</sub></i>)
 * = <i>c</i><sub>1</sub><i>x</i><sub>1</sub> + <i>c</i><sub>2</sub><i>x</i><sub>2</sub> + ... +
 * <i>c<sub>n</sub></i><i>x<sub>n</sub></i>, while <i>c</i><sub>1</sub>, <i>c</i><sub>2</sub>, ..., <i>c<sub>n</sub></i>
 * are constants.
 * <p>
 * Note 1: In the scheme, all element should be in some <i>Zp</i>, i.e. should be non-negative integers.
 * <p>
 * Note 2: Each participant has a unique ID, start from 0 to <i>n</i>-1.
 *
 * @author 		LoCCS
 * @version		1.0
 */

type LinearMultipartyComputation struct {
	/**
    * ID of this participant.
    */
	id int

	/**
	 * Number of participants.
	 */
	participantCount int

	/**
	 * Threshold <i>t</i>. Max number of semi-honest adversaries.
	 */
	threshold int

	/**
	 * <i>c</i><sub>1</sub>, <i>c</i><sub>2</sub>, ..., <i>c<sub>n</sub></i>. Coefficients of the linear function <i>f</i>.
	 */
	coefficients []interface{}

	/**
	 * Shamir's secret sharing scheme object.
	 */
	secretSharing secretshare.ShamirSecretSharingInterface

	/**
	 * The inputs received from other participants during the input stage.
	 */
	receivedInputs []interface{}

	/**
	 * The auxiliary data to generate Shamir's secret shares during the input stage.
	 */
	auxiliary []interface{}

	/**
	 * The outputs received from other participants during the output stage.
	 */
	receivedOutputs map[int]interface{}

	/**
    * Abstract Interfaces of LinearMultipartyComputation
    */
	linearMultipartyComputationCalculator LinearMultipartyComputationInterface
}

type LinearMultipartyComputationInterface interface {

	InitializeWithMaxValue(coefficients []interface{}, max interface{}) error

	InitializeWithModulus(coefficients []interface{}, modulus interface{}) error

	InitializeSimpleSumWithMax(max interface{}) error

	InitializeSimpleSumWithModulus(modulus interface{}) error

	GetModulus() interface{}

	GenerateInputAuxiliary() ([]interface{},error)

    GenerateInputs(secret interface{}, auxiliary []interface{}) ([]interface{},error)

	AddReceivedInput(from int, input interface{}) error

	HasAllInputReceived() bool

	GenerateOutput() (interface{}, error)

	AddReceivedOutput(from int, output interface{}) error

	isReadyForCompute() bool

	Compute() (interface{},error)

	Reset()

	/**
 	* Abstract method of getting a Shamir's secret sharing object with the number of participants and the modulus.
 	*
 	* @param participantCount The number of the participants.
 	* @param modulus The modulus <i>p</i>.
 	* @return The proper Shamir's secret sharing scheme object.
 	* @return error IllegalArgumentException If the number of participants or the modulus is invalid.
 	*/
 	getSecretSharing(participantCount int, modulus interface{}) (secretshare.ShamirSecretSharingInterface,error)

	/**
	* Abstract method of generating a proper modulus for Shamir's secret sharing from the coefficients of the linear function and the max value of the secret.
	*
	* @param coefficients The coefficients of the linear function.
	* @param max Max value of a secret.
	* @return The modulus for Shamir's secret sharing scheme.
	* @return error IllegalArgumentException If the coefficients or the max value is invalid.
	*/
	generateModulus(coefficients []interface{}, max interface{})(interface{},error)

	/**
 	* Abstract method of generating output during the output stage.
 	*
 	* @return The output value.
 	*/
	generateOutputImpl() interface{}

	/**
 	* Abstract method of checking if the coefficients and the modulus are both valid.
 	*
 	* @param coefficients Coefficients of the linear function.
 	* @param modulus Modulus of the Shamir's scheme.
 	* @return error if the coefficients or the modulus is invalid.
 	*/
	checkCoefficientsAndModulus(coefficients []interface{}, modulus interface{}) error

	/**
 	* Abstract method of getting an proper object for value 1.
 	*
 	* @return Element for value 1.
 	*/
 	getElementOne() interface{}

	/**
	* Abstract method of checking if the type of input element is valid.
	*
	* @param e Element to be checked.
	* @return True if the type of input element is valid, otherwise return false.
	*/
	checkElement(e interface{}) bool
}

/**
 * Set the linear function and try to find a proper modulus <i>p</i> by the max value of a secret.
 *
 * @param coefficients Coefficients of the linear function.
 * @param max Max value of a secret.
 * @return IllegalArgumentException If the coefficients or the max value is invalid.
 */
func (lmpc *LinearMultipartyComputation)InitializeWithMaxValue (coefficients []interface{}, max interface{}) error{
	if (coefficients == nil || len(coefficients) != lmpc.participantCount){
		return errors.New("Number of coefficients should be equal to number of participants.")
	}
	for i:=0; i < len(coefficients);i++{
		if (!lmpc.linearMultipartyComputationCalculator.checkElement(coefficients[i])){
			return errors.New("Invalid type of a coefficient.")
		}
	}
	if (!lmpc.linearMultipartyComputationCalculator.checkElement(max)){
		return errors.New("Invalid type of the max value of a secret.")
	}

	modulus,err := lmpc.linearMultipartyComputationCalculator.generateModulus(coefficients, max)
	if (err != nil) {return err}

	secretSharing, err := lmpc.linearMultipartyComputationCalculator.getSecretSharing(lmpc.participantCount, modulus)
    if (err != nil) {return err}
    lmpc.secretSharing = secretSharing

    access,err := secretshare.NewThresholdAccessStructure(lmpc.participantCount,lmpc.threshold)
	if (err != nil) {return err}
	lmpc.secretSharing.SetAccessStructure(access)

    lmpc.coefficients = coefficients
    return nil
}

/**
 * Set the linear function with coefficients and the modulus.
 *
 * @param coefficients Coefficients of the linear function.
 * @param modulus Modulus of the Shamir's scheme.
 * @return IllegalArgumentException If the coefficients or the modulus is invalid.
 */
func (lmpc *LinearMultipartyComputation) InitializeWithModulus(coefficients []interface{}, modulus interface{}) error{
	if (coefficients == nil || len(coefficients) != lmpc.participantCount){
		return errors.New("Number of coefficients should be equal to number of participants.")
	}
	for i:=0; i < len(coefficients);i++{
		if (!lmpc.linearMultipartyComputationCalculator.checkElement(coefficients[i])){
			return errors.New("Invalid type of a coefficient.")
		}
	}
	if (!lmpc.linearMultipartyComputationCalculator.checkElement(modulus)){
		return errors.New("Invalid type of modulus.")
	}

	err := lmpc.linearMultipartyComputationCalculator.checkCoefficientsAndModulus(coefficients,modulus)
	if (err != nil) {return err}

	secretSharing, err := lmpc.linearMultipartyComputationCalculator.getSecretSharing(lmpc.participantCount, modulus)
	if (err != nil) {return err}
	lmpc.secretSharing = secretSharing

	access,err := secretshare.NewThresholdAccessStructure(lmpc.participantCount,lmpc.threshold)
	if (err != nil) {return err}
	lmpc.secretSharing.SetAccessStructure(access)

	lmpc.coefficients = coefficients
	return nil
}

/**
 * Set a simple sum linear function, i.e. all values of the coefficients are 1, and
 * try to find a proper modulus <i>p</i> by the max value of a secret.
 *
 * @param max Max value of a secret.
 * @return IllegalArgumentException If the max value is invalid.
 */
func (lmpc *LinearMultipartyComputation) InitializeSimpleSumWithMax(max interface{}) error{
	if (!lmpc.linearMultipartyComputationCalculator.checkElement(max)){
		return errors.New("Invalid type of the max value of a secret.")
	}
    coefficients := make([]interface{}, lmpc.participantCount)
    for i := 0 ; i < len(coefficients) ; i++{
    	coefficients[i] = lmpc.linearMultipartyComputationCalculator.getElementOne()
	}
	modulus,err := lmpc.linearMultipartyComputationCalculator.generateModulus(coefficients, max)
	if (err != nil) {return err}

	secretSharing, err := lmpc.linearMultipartyComputationCalculator.getSecretSharing(lmpc.participantCount, modulus)
	if (err != nil) {return err}
	lmpc.secretSharing = secretSharing

	access,err := secretshare.NewThresholdAccessStructure(lmpc.participantCount,lmpc.threshold)
	if (err != nil) {return err}
	lmpc.secretSharing.SetAccessStructure(access)

	lmpc.coefficients = coefficients
	return nil
}

/**
 * Set a simple sum linear function, i.e. all values of the coefficients are 1, and the modulus.
 *
 * @param modulus Modulus of the Shamir's scheme.
 * @throws IllegalArgumentException If the modulus is invalid.
 */
func (lmpc *LinearMultipartyComputation) InitializeSimpleSumWithModulus(modulus interface{}) error{
	if (!lmpc.linearMultipartyComputationCalculator.checkElement(modulus)){
		return errors.New("Invalid type of modulus.")
	}
	coefficients := make([]interface{}, lmpc.participantCount)
	for i := 0 ; i < len(coefficients) ; i++{
		coefficients[i] = lmpc.linearMultipartyComputationCalculator.getElementOne()
	}

	err := lmpc.linearMultipartyComputationCalculator.checkCoefficientsAndModulus(coefficients,modulus)
	if (err != nil) {return err}

	secretSharing, err := lmpc.linearMultipartyComputationCalculator.getSecretSharing(lmpc.participantCount, modulus)
	if (err != nil) {return err}
	lmpc.secretSharing = secretSharing

	access,err := secretshare.NewThresholdAccessStructure(lmpc.participantCount,lmpc.threshold)
	if (err != nil) {return err}
	lmpc.secretSharing.SetAccessStructure(access)

	lmpc.coefficients = coefficients
	return nil
}

/**
 * Return the modulus in the Shamir's secret sharing scheme.
 *
 * @return The modulus <i>p</i>.
 */
func (lmpc *LinearMultipartyComputation)GetModulus() interface{}{
	return lmpc.secretSharing.GetModulus()
}

/**
 * Generate random auxiliary data in Shamir's scheme.
 *
 * @return Random auxiliary data.
 * @return error If Secret sharing scheme has not been set.
 */
func (lmpc *LinearMultipartyComputation) GenerateInputAuxiliary() ([]interface{},error){
	if (lmpc.secretSharing == nil){
		return nil, errors.New("Secret sharing scheme not set.")
	}
	return lmpc.secretSharing.GenerateRandomAuxiliary(),nil
}

/**
 * Generate inputs for all participants during the input stage.
 *
 * @param secret The secret value of this participant.
 * @param auxiliary The auxiliary data for generating Shamir's secret shares.
 * @return The inputs for all participants.
 * @return error IllegalArgumentException If the secret value or the auxiliary data is invalid.
 *         or IllegalStateException If the linear function or the secret sharing scheme in not set properly.
 */
func (lmpc *LinearMultipartyComputation) GenerateInputs(secret interface{}, auxiliary []interface{}) ([]interface{},error){
	if (lmpc.coefficients == nil || lmpc.secretSharing == nil){
		return nil, errors.New("Coefficients or secret sharing scheme not set.")
	}
	if (auxiliary == nil || len(auxiliary) != lmpc.participantCount){
		return nil, errors.New("Number of auxiliaries should be equal to number of participants.")
	}
	for i:=0; i < len(auxiliary);i++{
		if (!lmpc.linearMultipartyComputationCalculator.checkElement(auxiliary[i])){
			return nil, errors.New("Invalid type of an auxiliary.")
		}
	}
	if (!lmpc.linearMultipartyComputationCalculator.checkElement(secret)){
		return nil, errors.New("Invalid type a secret.")
	}

	shares,err := lmpc.secretSharing.GenerateShares(secret,auxiliary)
	if (err != nil) {return nil, err}

	inputs := make([]interface{}, lmpc.participantCount)
    for i := 0; i< lmpc.participantCount;i++{
    	inputs[i] = shares[i].GetValue().(*secretshare.ShamirSecretShareValue).GetQr()
	}
    lmpc.auxiliary = auxiliary
    lmpc.receivedInputs[lmpc.id] = inputs[lmpc.id] //itself
    return inputs, nil
}

/**
 * Add an input when received from other participant during the input stage.
 *
 * @param from The id of the participant who sent the input.
 * @param input The input value received.
 * @return error IllegalArgumentException If the id of the participant or the input value is invalid.
 */
func (lmpc *LinearMultipartyComputation) AddReceivedInput(from int, input interface{}) error{
	if ((from < 0) || (from >= lmpc.participantCount)){
		return errors.New("Invalid ID of the received input.")
	}
	if (!lmpc.linearMultipartyComputationCalculator.checkElement(input)){
		return errors.New("Invalid type of input.")
	}
	lmpc.receivedInputs[from] = input
	return nil
}

/**
 * Test if all <i>n</i>-1 inputs are received from other participants.
 *
 * @return True if all inputs are received, otherwise return false.
 */
func (lmpc *LinearMultipartyComputation) HasAllInputReceived() bool{
    for i := 0; i< lmpc.participantCount; i++{
    	if (lmpc.receivedInputs[i] == nil){
    		return false
		}
	}
	return true
}

/**
 * Generate the output during the output stage.
 * <p>
 * Call implemented <code>generateOutputImpl</code> to do the actually generating job.
 *
 * @return The output value.
 * @return error IllegalStateException If not all inputs are received or the secret sharing scheme in not set properly.
 */
func (lmpc *LinearMultipartyComputation) GenerateOutput() (interface{}, error){
	if (lmpc.coefficients == nil || lmpc.secretSharing == nil){
		return nil, errors.New("Coefficients or secret sharing scheme not set.")
	}
	if (!lmpc.HasAllInputReceived()){
		return nil, errors.New("Output cannot be generated before all inputs are received.")
	}
    output := lmpc.linearMultipartyComputationCalculator.generateOutputImpl()
    lmpc.receivedOutputs[lmpc.id] = output
    return output, nil
}

/**
 * Add an output when received from other participant during the output stage.
 *
 * @param from The id of the participant who sent the output.
 * @param output The output value received.
 * @return error IllegalArgumentException If the id of the participant or the output value is invalid.
 */
func (lmpc *LinearMultipartyComputation) AddReceivedOutput(from int, output interface{}) error{
	if ((from < 0) || (from >= lmpc.participantCount)){
		return errors.New("Invalid ID of the received output.")
	}
	if (!lmpc.linearMultipartyComputationCalculator.checkElement(output)){
		return errors.New("Invalid type of output.")
	}
	lmpc.receivedOutputs[from] = output
	return nil
}

/**
 * Test if enough outputs are received to compute the linear function.
 *
 * @return True if enough outputs are received, otherwise return false.
 */
func (lmpc *LinearMultipartyComputation) isReadyForCompute() bool{
    return len(lmpc.receivedOutputs) > lmpc.threshold
}

/**
 * Compute the linear function.
 *
 * @return The result value of the linear function. If some output value is wrong, return null.
 * @return error IllegalStateException If not enough outputs are received or the secret sharing scheme in not set properly.
 */
func (lmpc *LinearMultipartyComputation) Compute() (interface{},error){
	if (lmpc.secretSharing == nil){
		return nil,errors.New("Secret sharing scheme not inialized.")
	}
	if (lmpc.auxiliary == nil){
		return nil, errors.New("Secure MPC should start after generating input.")
	}
	if (!lmpc.isReadyForCompute()){
		return nil, errors.New("Not enough outputs received.")
	}
	shares := make([]*secretshare.SecretShare, lmpc.threshold+1)
	i := 0
	for k,v := range(lmpc.receivedOutputs){
		shareValue := secretshare.NewShamirSecretShareValue(lmpc.auxiliary[k], v)
		secretShare := secretshare.NewSecretShare(k,shareValue)
		shares[i] = secretShare
		i++
		if (i > lmpc.threshold) {break}
	}
	return lmpc.secretSharing.CalculateSecret(shares)
}

/**
 * Reset to time before input stage. And ready for the next round of MPC.
 */
func (lmpc *LinearMultipartyComputation) Reset(){
	if (lmpc.receivedInputs == nil) {return}
	for i := 0; i< len(lmpc.receivedInputs);i++{
		lmpc.receivedInputs[i] = nil
	}
	lmpc.auxiliary = nil
	lmpc.receivedOutputs = map[int]interface{} {}
}