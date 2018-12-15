package secretshare

import "errors"

/**
 * Abstract class for access structure in secret sharing schemes.
 *
 * <p>
 * The abstract class <code>AccessStructure</code> provides abstract method that tests
 * if a particular set of participants match this access structure. Subclasses of <code>AccessStructure</code>
 * should provide specific setting methods and implement the above testing method.
 * <p>
 * Echo participant in the secret sharing system has an unique int ID, starting from 0 to (N - 1) (N is the
 * number of participants).
 *
 * @author 		LoCCS
 * @version		1.0
 *
 */
type AccessStructure struct {
	/**
    * The number of participants that share the secret.
    */
    participantCount int

	/**
    * Abstract Interfaces of AccessStructure
	*/
	AccessStructureITF AccessStructureInterface
}

type AccessStructureInterface interface {

	GetParticipantCount() int

    TestAccessable (participants []int) (bool,error)

	GetThreshold() int

	/**
    * Abstract method of testing whether a particular set of participants match this access structure.
    *
    * @param participants IDs of the participants to be tested.
    * @return If the participants match this access structure.
    */
	testAccessableImpl(participants []int) (bool)
}

/**
* Get the number of participants that share the secret.
*
* @return Number of participants that share the secret.
*/
func (accs *AccessStructure) GetParticipantCount() int{
	return accs.participantCount
}

/**
* Test if a particular set of participants match this access structure.
* <p>
* Call implemented <code>testAccessableImpl</code> method to do the testing job in subclasses.
*
* @param participants IDs of the participants to be tested.
* @return If the participants match this access structure.
* @return error If any ID of the input participants is invalid.
*/
func (accs *AccessStructure) TestAccessable(participants []int) (bool,error){
	for i := 0; i < len(participants); i++ {
	    if ((participants[i] < 0) || (participants[i] >= accs.participantCount)){
	    	return false, errors.New("Invalid participant ID.");
		}
	}
	return accs.AccessStructureITF.testAccessableImpl(participants), nil;
}
