// Copyright 2013 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestPullRequestsService_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"state":     "closed",
			"head":      "h",
			"base":      "b",
			"sort":      "created",
			"direction": "desc",
			"page":      "2",
		})
		fmt.Fprint(w, `[{"number":1}]`)
	})

	opt := &PullRequestListOptions{"closed", "h", "b", "created", "desc", ListOptions{Page: 2}}
	pulls, _, err := client.PullRequests.List("o", "r", opt)
	if err != nil {
		t.Errorf("PullRequests.List returned error: %v", err)
	}

	want := []*PullRequest{{Number: Int(1)}}
	if !reflect.DeepEqual(pulls, want) {
		t.Errorf("PullRequests.List returned %+v, want %+v", pulls, want)
	}
}

func TestPullRequestsService_List_invalidOwner(t *testing.T) {
	_, _, err := client.PullRequests.List("%", "r", nil)
	testURLParseError(t, err)
}

func TestPullRequestsService_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"number":1}`)
	})

	pull, _, err := client.PullRequests.Get("o", "r", 1)
	if err != nil {
		t.Errorf("PullRequests.Get returned error: %v", err)
	}

	want := &PullRequest{Number: Int(1)}
	if !reflect.DeepEqual(pull, want) {
		t.Errorf("PullRequests.Get returned %+v, want %+v", pull, want)
	}
}

func TestPullRequestService_GetRawDiff(t *testing.T) {
	setup()
	defer teardown()
	const rawStr = "@@diff content"

	mux.HandleFunc("/repos/o/r/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeV3Diff)
		fmt.Fprint(w, rawStr)
	})

	ret, _, err := client.PullRequests.GetRaw("o", "r", 1, RawOptions{Diff})
	if err != nil {
		t.Fatalf("PullRequests.GetRaw returned error: %v", err)
	}

	if ret != rawStr {
		t.Errorf("PullRequests.GetRaw returned %s want %s", ret, rawStr)
	}
}

func TestPullRequestService_GetRawPatch(t *testing.T) {
	setup()
	defer teardown()
	const rawStr = "@@patch content"

	mux.HandleFunc("/repos/o/r/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeV3Patch)
		fmt.Fprint(w, rawStr)
	})

	ret, _, err := client.PullRequests.GetRaw("o", "r", 1, RawOptions{Patch})
	if err != nil {
		t.Fatalf("PullRequests.GetRaw returned error: %v", err)
	}

	if ret != rawStr {
		t.Errorf("PullRequests.GetRaw returned %s want %s", ret, rawStr)
	}
}

func TestPullRequestService_GetRawInvalid(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.PullRequests.GetRaw("o", "r", 1, RawOptions{100})
	if err == nil {
		t.Fatal("PullRequests.GetRaw should return error")
	}
	if !strings.Contains(err.Error(), "unsupported raw type") {
		t.Error("PullRequests.GetRaw should return unsupported raw type error")
	}
}

func TestPullRequestsService_Get_headAndBase(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"number":1,"head":{"ref":"r2","repo":{"id":2}},"base":{"ref":"r1","repo":{"id":1}}}`)
	})

	pull, _, err := client.PullRequests.Get("o", "r", 1)
	if err != nil {
		t.Errorf("PullRequests.Get returned error: %v", err)
	}

	want := &PullRequest{
		Number: Int(1),
		Head: &PullRequestBranch{
			Ref:  String("r2"),
			Repo: &Repository{ID: Int(2)},
		},
		Base: &PullRequestBranch{
			Ref:  String("r1"),
			Repo: &Repository{ID: Int(1)},
		},
	}
	if !reflect.DeepEqual(pull, want) {
		t.Errorf("PullRequests.Get returned %+v, want %+v", pull, want)
	}
}

func TestPullRequestsService_Get_urlFields(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"number":1,
			"url": "https://api.github.com/repos/octocat/Hello-World/pulls/1347",
			"html_url": "https://github.com/octocat/Hello-World/pull/1347",
			"issue_url": "https://api.github.com/repos/octocat/Hello-World/issues/1347",
			"statuses_url": "https://api.github.com/repos/octocat/Hello-World/statuses/6dcb09b5b57875f334f61aebed695e2e4193db5e",
			"diff_url": "https://github.com/octocat/Hello-World/pull/1347.diff",
			"patch_url": "https://github.com/octocat/Hello-World/pull/1347.patch",
			"review_comments_url": "https://api.github.com/repos/octocat/Hello-World/pulls/1347/comments",
			"review_comment_url": "https://api.github.com/repos/octocat/Hello-World/pulls/comments{/number}"}`)
	})

	pull, _, err := client.PullRequests.Get("o", "r", 1)
	if err != nil {
		t.Errorf("PullRequests.Get returned error: %v", err)
	}

	want := &PullRequest{
		Number:            Int(1),
		URL:               String("https://api.github.com/repos/octocat/Hello-World/pulls/1347"),
		HTMLURL:           String("https://github.com/octocat/Hello-World/pull/1347"),
		IssueURL:          String("https://api.github.com/repos/octocat/Hello-World/issues/1347"),
		StatusesURL:       String("https://api.github.com/repos/octocat/Hello-World/statuses/6dcb09b5b57875f334f61aebed695e2e4193db5e"),
		DiffURL:           String("https://github.com/octocat/Hello-World/pull/1347.diff"),
		PatchURL:          String("https://github.com/octocat/Hello-World/pull/1347.patch"),
		ReviewCommentsURL: String("https://api.github.com/repos/octocat/Hello-World/pulls/1347/comments"),
		ReviewCommentURL:  String("https://api.github.com/repos/octocat/Hello-World/pulls/comments{/number}"),
	}

	if !reflect.DeepEqual(pull, want) {
		t.Errorf("PullRequests.Get returned %+v, want %+v", pull, want)
	}
}

func TestPullRequestsService_Get_invalidOwner(t *testing.T) {
	_, _, err := client.PullRequests.Get("%", "r", 1)
	testURLParseError(t, err)
}

func TestPullRequestsService_Create(t *testing.T) {
	setup()
	defer teardown()

	input := &NewPullRequest{Title: String("t")}

	mux.HandleFunc("/repos/o/r/pulls", func(w http.ResponseWriter, r *http.Request) {
		v := new(NewPullRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"number":1}`)
	})

	pull, _, err := client.PullRequests.Create("o", "r", input)
	if err != nil {
		t.Errorf("PullRequests.Create returned error: %v", err)
	}

	want := &PullRequest{Number: Int(1)}
	if !reflect.DeepEqual(pull, want) {
		t.Errorf("PullRequests.Create returned %+v, want %+v", pull, want)
	}
}

func TestPullRequestsService_Create_invalidOwner(t *testing.T) {
	_, _, err := client.PullRequests.Create("%", "r", nil)
	testURLParseError(t, err)
}

func TestPullRequestsService_Edit(t *testing.T) {
	setup()
	defer teardown()

	input := &PullRequest{Title: String("t")}

	mux.HandleFunc("/repos/o/r/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		v := new(PullRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "PATCH")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"number":1}`)
	})

	pull, _, err := client.PullRequests.Edit("o", "r", 1, input)
	if err != nil {
		t.Errorf("PullRequests.Edit returned error: %v", err)
	}

	want := &PullRequest{Number: Int(1)}
	if !reflect.DeepEqual(pull, want) {
		t.Errorf("PullRequests.Edit returned %+v, want %+v", pull, want)
	}
}

func TestPullRequestsService_Edit_invalidOwner(t *testing.T) {
	_, _, err := client.PullRequests.Edit("%", "r", 1, nil)
	testURLParseError(t, err)
}

func TestPullRequestsService_ListCommits(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/1/commits", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{"page": "2"})
		fmt.Fprint(w, `
			[
			  {
			    "sha": "3",
			    "parents": [
			      {
			        "sha": "2"
			      }
			    ]
			  },
			  {
			    "sha": "2",
			    "parents": [
			      {
			        "sha": "1"
			      }
			    ]
			  }
			]`)
	})

	opt := &ListOptions{Page: 2}
	commits, _, err := client.PullRequests.ListCommits("o", "r", 1, opt)
	if err != nil {
		t.Errorf("PullRequests.ListCommits returned error: %v", err)
	}

	want := []*RepositoryCommit{
		{
			SHA: String("3"),
			Parents: []Commit{
				{
					SHA: String("2"),
				},
			},
		},
		{
			SHA: String("2"),
			Parents: []Commit{
				{
					SHA: String("1"),
				},
			},
		},
	}
	if !reflect.DeepEqual(commits, want) {
		t.Errorf("PullRequests.ListCommits returned %+v, want %+v", commits, want)
	}
}

func TestPullRequestsService_ListFiles(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/1/files", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{"page": "2"})
		fmt.Fprint(w, `
			[
			  {
			    "sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			    "filename": "file1.txt",
			    "status": "added",
			    "additions": 103,
			    "deletions": 21,
			    "changes": 124,
			    "patch": "@@ -132,7 +132,7 @@ module Test @@ -1000,7 +1000,7 @@ module Test"
			  },
			  {
			    "sha": "f61aebed695e2e4193db5e6dcb09b5b57875f334",
			    "filename": "file2.txt",
			    "status": "modified",
			    "additions": 5,
			    "deletions": 3,
			    "changes": 103,
			    "patch": "@@ -132,7 +132,7 @@ module Test @@ -1000,7 +1000,7 @@ module Test"
			  }
			]`)
	})

	opt := &ListOptions{Page: 2}
	commitFiles, _, err := client.PullRequests.ListFiles("o", "r", 1, opt)
	if err != nil {
		t.Errorf("PullRequests.ListFiles returned error: %v", err)
	}

	want := []*CommitFile{
		{
			SHA:       String("6dcb09b5b57875f334f61aebed695e2e4193db5e"),
			Filename:  String("file1.txt"),
			Additions: Int(103),
			Deletions: Int(21),
			Changes:   Int(124),
			Status:    String("added"),
			Patch:     String("@@ -132,7 +132,7 @@ module Test @@ -1000,7 +1000,7 @@ module Test"),
		},
		{
			SHA:       String("f61aebed695e2e4193db5e6dcb09b5b57875f334"),
			Filename:  String("file2.txt"),
			Additions: Int(5),
			Deletions: Int(3),
			Changes:   Int(103),
			Status:    String("modified"),
			Patch:     String("@@ -132,7 +132,7 @@ module Test @@ -1000,7 +1000,7 @@ module Test"),
		},
	}

	if !reflect.DeepEqual(commitFiles, want) {
		t.Errorf("PullRequests.ListFiles returned %+v, want %+v", commitFiles, want)
	}
}

func TestPullRequestsService_IsMerged(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/1/merge", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusNoContent)
	})

	isMerged, _, err := client.PullRequests.IsMerged("o", "r", 1)
	if err != nil {
		t.Errorf("PullRequests.IsMerged returned error: %v", err)
	}

	want := true
	if !reflect.DeepEqual(isMerged, want) {
		t.Errorf("PullRequests.IsMerged returned %+v, want %+v", isMerged, want)
	}
}

func TestPullRequestsService_Merge(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/1/merge", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		testHeader(t, r, "Accept", mediaTypeSquashPreview)
		fmt.Fprint(w, `
			{
			  "sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			  "merged": true,
			  "message": "Pull Request successfully merged"
			}`)
	})

	options := &PullRequestOptions{MergeMethod: "rebase"}
	merge, _, err := client.PullRequests.Merge("o", "r", 1, "merging pull request", options)
	if err != nil {
		t.Errorf("PullRequests.Merge returned error: %v", err)
	}

	want := &PullRequestMergeResult{
		SHA:     String("6dcb09b5b57875f334f61aebed695e2e4193db5e"),
		Merged:  Bool(true),
		Message: String("Pull Request successfully merged"),
	}
	if !reflect.DeepEqual(merge, want) {
		t.Errorf("PullRequests.Merge returned %+v, want %+v", merge, want)
	}
}
