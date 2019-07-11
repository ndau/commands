#!/usr/bin/env node

yaml = require('js-yaml');
fs   = require('fs');

async function main (){
    const FILE = process.argv[2]
    const BRANCH = process.argv[3] || ''
    const TAG = process.argv[4] || ''

    // Get document, or throw exception on error
    try {
        var doc = yaml.safeLoad(fs.readFileSync(FILE, 'utf8'));
    Object.keys(doc.workflows).forEach((k)=>{
        if (k==="version" || k===undefined) { return ; }
        const w = doc.workflows[k]
        console.log(`Workflow: ${k}`)
        w.jobs.forEach((j)=>{
            const name = Object.keys(j)[0]
            let oB = true
            let iB = true
            let oT = true
            let iT = true

            onlyTag   = (((j[name].filters||{tags:{}}).tags||{}).only  ||'').replace(/\//g,'')
            ignoreTag = (((j[name].filters||{tags:{}}).tags||{}).ignore||'').replace(/\//g,'')
            if (onlyTag) {
                const re = new RegExp(onlyTag)
                oT = Boolean(TAG.match(re))
            }
            if (ignoreTag) {
                const re = new RegExp(ignoreTag)
                iT = Boolean(TAG.match(re))
            }

            onlyBranch   = (((j[name].filters||{branches:{}}).branches||{}).only  ||'').replace(/\//g,'')
            ignoreBranch = (((j[name].filters||{branches:{}}).branches||{}).ignore||'').replace(/\//g,'')
            if (onlyBranch) {
                const re = new RegExp(onlyBranch)
                oB = Boolean(BRANCH.match(re))
            }
            if (ignoreBranch) {
                const re = new RegExp(ignoreBranch)
                iB = Boolean(BRANCH.match(re))
            }
            if ((!iB && !iT)) {
                console.log(` - ${name}`)
            }
            if ((onlyBranch && oB) || (onlyTag && oT)) {
                console.log(` - ${name}`)
            }
        })
    })
    } catch (e) {
      console.log(e)
    }
}

main()

