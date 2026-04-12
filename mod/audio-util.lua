function amp2db(amp)
    return 20 * math.log(amp, 10)
end

function db2amp(db)
    return 10 ^ (db / 20)
end
