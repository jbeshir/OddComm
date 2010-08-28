/*
	Implements hookable permission and validation checks.

	For all types of permission, hooks return an integer to indicate
	whether they grant permission or not, and the level of the permission
	granted/restriction imposed.

	Standard levels are as follows:

	0: Do not modify permissions.

	1/-1: Default permit/deny. Should not be returned by hooks.

	100/-100: Permit/deny normal users.

	10000/-10000: Permit/deny at the channel operator level.

	1000000/-1000000: Permit/deny at the network operator level.

	-1000000000: Always deny, not overridable. For use when this permission
	being granted would cause the server to enter an invalid state. Note
	that only validation checks are required to be performed, and this is
	the only result required to be honoured from them.

	The greatest magnitude result is used, with a positive result
	overriding a negative result of the same level.
*/
package perm
