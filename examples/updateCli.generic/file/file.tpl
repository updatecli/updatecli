---
source:
  kind: file
  spec:
    # Test url
    file: https://get.helm.sh/helm-v3.5.0-darwin-amd64.tar.gz.sha256sum
    # Test file url
    #file: file://LICENSE
    # Test file
    # file: LICENSE
  scm:
    github:
      user: "{{ .github.user }}"
      email: "{{ .github.email }}"
      owner: "{{ .github.owner }}"
      repository: "{{ .github.repository }}"
      token: "{{ requiredEnv .github.token }}"
      username: "{{ .github.username }}"
      branch: "{{ .github.branch }}"
conditions:
  condition0:
    name: condition0
    kind: file
    spec:
      file: CODEOWNERS
  condition1:
    name: condition1
    kind: file
    spec:
      file: LICENSE
  condition2:
    name: condition2
    kind: file
    spec:
      file: LICENSE
      # !!! \n is important at the end of the file
      content: |
        MIT License
        
        Copyright (c) 2020 Olblak
        
        Permission is hereby granted, free of charge, to any person obtaining a copy
        of this software and associated documentation files (the "Software"), to deal
        in the Software without restriction, including without limitation the rights
        to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
        copies of the Software, and to permit persons to whom the Software is
        furnished to do so, subject to the following conditions:
        
        The above copyright notice and this permission notice shall be included in all
        copies or substantial portions of the Software.
        
        THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
        IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
        FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
        AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
        LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
        OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
        SOFTWARE.
    scm:
      github:
        user: "{{ .github.user }}"
        email: "{{ .github.email }}"
        owner: "{{ .github.owner }}"
        repository: "{{ .github.repository }}"
        token: "{{ requiredEnv .github.token }}"
        username: "{{ .github.username }}"
        branch: "{{ .github.branch }}"
  condition3:
    name: condition3
    kind: file
    spec:
      file: https://get.helm.sh/helm-v3.5.0-darwin-amd64.tar.gz.sha256sum
      content: |
        53378d8de087395ece3876795111a8077f2451f080ec6150d777cc3105214520  helm-v3.5.0-darwin-amd64.tar.gz
  condition4:
    name: condition4
    kind: file
    spec:
      file: /home/olblak/Project/Olblak/updatecli/LICENSE
      content: |
        MIT License
        
        Copyright (c) 2020 Olblak
        
        Permission is hereby granted, free of charge, to any person obtaining a copy
        of this software and associated documentation files (the "Software"), to deal
        in the Software without restriction, including without limitation the rights
        to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
        copies of the Software, and to permit persons to whom the Software is
        furnished to do so, subject to the following conditions:
        
        The above copyright notice and this permission notice shall be included in all
        copies or substantial portions of the Software.
        
        THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
        IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
        FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
        AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
        LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
        OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
        SOFTWARE.
targets:
  file1:
    name: target1
    kind: file
    spec:
      file: LICENSE
    scm:
      github:
        user: "{{ .github.user }}"
        email: "{{ .github.email }}"
        owner: "{{ .github.owner }}"
        repository: "{{ .github.repository }}"
        token: "{{ requiredEnv .github.token }}"
        username: "{{ .github.username }}"
        branch: "{{ .github.branch }}"
  file2:
    name: target2
    kind: file
    spec:
      file: CODEOWNERS
    scm:
      github:
        user: "{{ .github.user }}"
        email: "{{ .github.email }}"
        owner: "{{ .github.owner }}"
        repository: "{{ .github.repository }}"
        token: "{{ requiredEnv .github.token }}"
        username: "{{ .github.username }}"
        branch: "{{ .github.branch }}"
