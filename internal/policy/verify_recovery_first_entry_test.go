// Copyright The gittuf Authors
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"testing"

	"github.com/gittuf/gittuf/internal/common"
	"github.com/gittuf/gittuf/internal/rsl"
	"github.com/gittuf/gittuf/pkg/gitinterface"
	"github.com/stretchr/testify/assert"
)

func TestVerifyRelativeForRefFirstEntryInvalid(t *testing.T) {
	repo, _ := createTestRepository(t, createTestStateWithPolicy)
	refName := "refs/heads/main"

	// 1. Push an unauthorized commit (Entry 1)
	commitIDs := common.AddNTestCommitsToSpecifiedRef(t, repo, refName, 1, gpgUnauthorizedKeyBytes)
	entry1 := rsl.NewReferenceEntry(refName, commitIDs[0])
	entry1ID := common.CreateTestRSLReferenceEntryCommit(t, repo, entry1, gpgUnauthorizedKeyBytes)
	entry1.ID = entry1ID

	// 2. Add a skip annotation for Entry 1 (Entry 2)
	annotation := rsl.NewAnnotationEntry([]gitinterface.Hash{entry1ID}, true, "invalid entry")
	common.CreateTestRSLAnnotationEntryCommit(t, repo, annotation, gpgKeyBytes)

	// 3. Push an authorized commit (Entry 3) as a new root commit
	// We delete the ref first to ensure CommitUsingSpecificKey doesn't set a parent
	err := repo.CheckAndSetReference(refName, gitinterface.ZeroHash, commitIDs[0])
	assert.Nil(t, err)

	emptyTreeID, err := repo.EmptyTree()
	assert.Nil(t, err)
	commit3ID, err := repo.CommitUsingSpecificKey(emptyTreeID, refName, "Authorized first commit", gpgKeyBytes)
	assert.Nil(t, err)
	entry3 := rsl.NewReferenceEntry(refName, commit3ID)
	entry3ID := common.CreateTestRSLReferenceEntryCommit(t, repo, entry3, gpgKeyBytes)
	entry3.ID = entry3ID

	verifier := NewPolicyVerifier(repo)

	// This is where it currently fails because Entry 1 is invalid and skipped,
	// but it's the first entry for the ref.
	err = verifier.VerifyRelativeForRef(testCtx, entry1, entry3, refName)
	assert.Nil(t, err)
}

func TestVerifyRelativeForRefFirstEntryInvalidNoFix(t *testing.T) {
	repo, _ := createTestRepository(t, createTestStateWithPolicy)
	refName := "refs/heads/main"

	// 1. Push an unauthorized commit (Entry 1)
	commitIDs := common.AddNTestCommitsToSpecifiedRef(t, repo, refName, 1, gpgUnauthorizedKeyBytes)
	entry1 := rsl.NewReferenceEntry(refName, commitIDs[0])
	entry1ID := common.CreateTestRSLReferenceEntryCommit(t, repo, entry1, gpgUnauthorizedKeyBytes)
	entry1.ID = entry1ID

	// 2. Add a skip annotation for Entry 1 (Entry 2)
	annotation := rsl.NewAnnotationEntry([]gitinterface.Hash{entry1ID}, true, "invalid entry")
	common.CreateTestRSLAnnotationEntryCommit(t, repo, annotation, gpgKeyBytes)

	// 3. Push an authorized commit (Entry 3) that is NOT tree-same with the empty tree
	// We ensure it has at least one file so its tree is not empty
	commit3IDs := common.AddNTestCommitsToSpecifiedRef(t, repo, refName, 1, gpgKeyBytes)
	entry3 := rsl.NewReferenceEntry(refName, commit3IDs[0])
	entry3ID := common.CreateTestRSLReferenceEntryCommit(t, repo, entry3, gpgKeyBytes)
	entry3.ID = entry3ID

	verifier := NewPolicyVerifier(repo)

	// This should now fail because entry3 is not a fix for entry1
	err := verifier.VerifyRelativeForRef(testCtx, entry1, entry3, refName)
	assert.ErrorIs(t, err, ErrFirstRSLEntryInvalidUnrecoverable)
}
