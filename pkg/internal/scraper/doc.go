// Copyright © 2020 Elis Lulja
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package scraper contains code that scrapes the website.
// This package is meant to be a singleton: you can use this with multiple
// pages of the same website as long as the pages are similar.
//
// You should run different instances of the program with a different
// implementation of `Scrape` if you want to scrape pages that have a different
// html/body structure.
//
// For example, if pages A, B, C all have a `<div id='title'>` you can
// implement `Scrape` to read that html even if A, B, C are from different
// websites. Then you build and deploy this as a pod.
// But if pages D, E have `<div id='main-title'>` then you should create a
// different implementation and deploy this as a different pod.
//
// You can also just do a `switch` case and do everything on one instance
// of `Scrape`, but this defeats the purpose of the program, as this is made
// so that you don't have to update all pods, but just the ones that need it,
// and let the others continue their work.
package scraper
