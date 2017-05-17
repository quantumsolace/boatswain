# boatswain

## Config file lives in `${HOME}/.boatswain.yaml`

## `boatswain stage push`
In order to build and deploy automatically to staging using `boatswain stage push`, you need to add a `builds` array in your config with an entry for each project you want to deploy to staging along with an absolute path to the project's root directory (`rootpath`) and the relative path — from the project's root — to the `build.sh` script (`path`). We need both values because `docker build` needs to be run from the project's root directory, but we also need to know where the `build.sh` script lives. See the example config below.

```yaml
release: /path/to/boatswain-values/deployment
builds:
- name: medbridge
  path: deployment/build.sh
  rootpath: /Users/<name>/Programming/Php/Medbridge
- name: medflix
  path: deployment/build.sh
  rootpath: /Users/<name>/Programming/Php/Medflix/
- name: ace
  path: deployment/build.sh
  rootpath: /Users/<name>/Programming/Php/Ace
```
