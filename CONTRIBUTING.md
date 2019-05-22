1. Install [Go](https://golang.org/doc/install)
2. Install [Git](https://help.github.com/en/articles/set-up-git)
3. Fork this repo, it will create a copy of this repo in your account
4. Clone forked repo into your GOPATH
```
git clone https://github.com/your-username/Ansiblock.git
```
5. Pick an issue you want to contribute to.
6. Create a branch with the id of an issue
```
git checkout -b ANS-1234
```
7. Make nessesary changes
8. Commit changes. Be sure to [write good commit messages](https://chris.beams.io/posts/git-commit)
9. Squash your commits into a single commit with git's [interactive rebase](https://www.atlassian.com/git/tutorials/rewriting-history/git-rebase)
10. Push your changes to your fork on Github
```
git push origin ANS-1234
``` 
11. From your fork open a pull request to `master`
12. Once the pull request is approved and merged you can pull the changes from `upstream` to your local repo and delete your extra branch.
