const config = {
    verbose: true,
    testEnvironment: 'jsdom',
    transform: {
        "^.+\\.js?$": "esbuild-jest"
    }
};

module.exports = config;
