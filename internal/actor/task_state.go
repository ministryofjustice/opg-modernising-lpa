package actor

import "github.com/ministryofjustice/opg-modernising-lpa/internal/task"

type TaskState = task.State

const (
	TaskNotStarted = task.NotStarted
	TaskInProgress = task.InProgress
	TaskCompleted  = task.Completed
)

type PaymentTask = task.PaymentState

const (
	PaymentTaskNotStarted           = task.PaymentStateNotStarted
	PaymentTaskInProgress           = task.PaymentStateInProgress
	PaymentTaskPending              = task.PaymentStatePending
	PaymentTaskApproved             = task.PaymentStateApproved
	PaymentTaskDenied               = task.PaymentStateDenied
	PaymentTaskMoreEvidenceRequired = task.PaymentStateMoreEvidenceRequired
	PaymentTaskCompleted            = task.PaymentStateCompleted
)

type IdentityTask = task.IdentityState

const (
	IdentityTaskNotStarted = task.IdentityStateNotStarted
	IdentityTaskInProgress = task.IdentityStateInProgress
	IdentityTaskPending    = task.IdentityStatePending
	IdentityTaskProblem    = task.IdentityStateProblem
	IdentityTaskCompleted  = task.IdentityStateCompleted
)
