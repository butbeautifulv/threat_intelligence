import (
    "./schema/lola.cue"
    "./schema/vuln.cue"
    "./schema/ti.cue"
    ds "./schema/ds.cue"
)

schema: {
    lola:   lola.lola
    vuln:   vuln.vuln
    ti:     ti.ti
    detect: ds.detect
}
