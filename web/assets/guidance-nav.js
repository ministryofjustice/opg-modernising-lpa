export class GuidanceNav {
    init() {
        const backToTopLinkWrapper = document.getElementById("back-to-top-link-wrapper")
        const guidanceNav = document.getElementById("guidance-nav")
        const guidanceContent = document.getElementById("guidance-content")

        if (guidanceNav && guidanceContent && backToTopLinkWrapper) {
            backToTopLinkWrapper.style.minHeight = guidanceContent.offsetHeight - guidanceNav.offsetHeight + "px"
        }
    }
}
