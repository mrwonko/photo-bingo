package main

const (
	basePath = "/photo-bingo"
	verbose  = true

	latestStatePath   = "state.json"
	previousStatePath = "state.prev.json"
)

var options = [5*5 - 1]Goal{
	{Name: "Shadows", Description: "Make shadows the main subject of your photograph. Focus on their shapes, patterns, and how they define or obscure light."},
	{Name: "Mostly dark", Description: "Create a high-contrast image where the majority of the frame is dark (low-key). Focus on using small amounts of light to highlight your subject."},
	{Name: "Lots of little details", Description: "Get up close and focus on the small, often-overlooked details of objects, textures, or nature. Think macro photography."},
	{Name: "Get low", Description: "Change your perspective. Crouch down, kneel, or even lie on the ground to capture a low-angle shot. What new details or drama appear when you photograph from below?"},
	{Name: "Panoramic", Description: "Capture a wide, expansive scene. This could be a sweeping landscape or an interesting interior. Experiment with your phone or camera's panoramic feature."},
	{Name: "Tilted Frame (Dutch Angle)", Description: "Intentionally angle or tilt your camera to create a dynamic, diagonal horizon line. This technique adds an unstable, uneasy, or exciting mood to the composition."},
	{Name: "Motion Blur", Description: "Convey the speed of a moving subject by following its movement with the camera at a low shutter speed to create motion blur in the background, or keep the camera fixed and let the subject become blurry"},
	{Name: "Self portrait", Description: "Create a photo where you are the subject. This doesn't have to be a traditional selfie; it could be a conceptual shot, a photo of your shadow, or a reflection."},
	{Name: "A single color", Description: "Choose a dominant color and find subjects or scenes where that color is the main visual element. Focus on its various shades and tones."},
	{Name: "Color-POW!!!", Description: "Create a photograph dominated by vibrant, saturated colors. Look for bold color combinations or striking pops of color against a neutral background."},
	{Name: "Monochrome", Description: "Shoot for a black and white image. Focus on texture, light, shadow, and composition, as color will not be a factor."},
	{Name: "Reflections", Description: "Look for reflections in water, windows, mirrors, or any shiny surface. Play with symmetry, distortion, and how the reflected world interacts with the real one."},
	{Name: "Typography", Description: "Find interesting examples of letters, numbers, signs, or graffiti. Focus on the style, texture, and context of the text you photograph."},
	{Name: "Street portrait", Description: "Capture a portrait of a person you encounter during the photowalk. Remember to ask for permission! Focus on expression, context, and environment."},
	{Name: "History", Description: "Find a subject that tells a story of the past. This could be old architecture, a historical marker, or an object with a sense of age."},
	{Name: "Look behind", Description: "Turn around and photograph something that is usually behind you or overlooked. Capture the view in the opposite direction of where you're walking."},
	{Name: "History", Description: "Find a subject that tells a story of the past. This could be old architecture, a historical marker, or an object with a sense of age."},
	{Name: "Look behind", Description: "Turn around and photograph something that is usually behind you or overlooked. Capture the view in the opposite direction of where you're walking."},
	{Name: "Abstract", Description: "Capture an image where the subject is obscured or non-representational. Focus on shapes, lines, textures, and colors to create a piece of visual art."},
	{Name: "Geometry", Description: "Focus on lines, shapes, patterns, and forms in your environment. Look for triangles, circles, squares, leading lines, or repeating elements to create a strong, structured composition."},
	{Name: "Leading Lines", Description: "Find a composition where lines (roads, fences, shadows, railings) naturally guide the viewer's eye toward your main subject or deeper into the frame. Focus on creating depth and visual flow."},
	{Name: "", Description: ""},
	{Name: "", Description: ""},
	{Name: "", Description: ""},
}
