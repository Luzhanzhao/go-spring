/*
 * Copyright 2012-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package EchoStarter

import (
	"github.com/go-spring/go-spring-web/spring-echo"
	"github.com/go-spring/go-spring-web/spring-web"
	"github.com/go-spring/go-spring/spring-boot"
	"github.com/go-spring/go-spring/spring-core"
	"github.com/go-spring/go-spring/starter-web"
)

func init() {

	SpringBoot.RegisterNameBeanFn("echo-web-container", func(config WebStarter.WebServerConfig) SpringWeb.WebContainer {
		return SpringEcho.NewContainer(SpringWeb.ContainerConfig{
			Port: config.Port,
		})
	}).ConditionOnPropertyValue("web.server.enable", true, SpringCore.MatchIfMissing(true))

	SpringBoot.RegisterNameBeanFn("echo-ssl-web-container", func(config WebStarter.WebServerConfig) SpringWeb.WebContainer {
		return SpringEcho.NewContainer(SpringWeb.ContainerConfig{
			EnableSSL: true,
			Port:      config.SSLPort,
			KeyFile:   config.SSLKey,
			CertFile:  config.SSLCert,
		})
	}).ConditionOnPropertyValue("web.server.ssl.enable", true)
}
