// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// An app that draws a green triangle on a red background.
//
// Note: This demo is an early preview of Go 1.5. In order to build this
// program as an Android APK using the gomobile tool.
//
// See http://godoc.org/golang.org/x/mobile/cmd/gomobile to install gomobile.
//
// Get the basic example and use gomobile to build or install it on your device.
//
//   $ go get -d golang.org/x/mobile/example/basic
//   $ gomobile build golang.org/x/mobile/example/basic # will build an APK
//
//   # plug your Android device to your computer or start an Android emulator.
//   # if you have adb installed on your machine, use gomobile install to
//   # build and deploy the APK to an Android target.
//   $ gomobile install golang.org/x/mobile/example/basic
//
// Switch to your device or emulator to start the Basic application from
// the launcher.
// You can also run the application on your desktop by running the command
// below. (Note: It currently doesn't work on Windows.)
//   $ go install golang.org/x/mobile/example/basic && basic
package main

import (
    "encoding/binary"
    "log"

    "golang.org/x/mobile/app"
    "golang.org/x/mobile/event/size"
    "golang.org/x/mobile/event/lifecycle"
    "golang.org/x/mobile/event/paint"
    "golang.org/x/mobile/event/touch"
    "golang.org/x/mobile/exp/app/debug"
    "golang.org/x/mobile/exp/f32"
    "golang.org/x/mobile/exp/gl/glutil"
    "golang.org/x/mobile/geom"
    "golang.org/x/mobile/gl"
)

var (
    program gl.Program
    position gl.Attrib
    scan gl.Uniform
    color gl.Uniform
    buf gl.Buffer

    touchLoc geom.Point
)

func main() {
    app.Main(func(a app.App) {
        var c size.Event
        for e := range a.Events() {
            switch e := app.Filter(e).(type) {
                case lifecycle.Event:
                switch e.Crosses(lifecycle.StageVisible) {
                    case lifecycle.CrossOn:
                    onStart()
                    case lifecycle.CrossOff:
                    onStop()
                }
                case size.Event:
                c = e
                touchLoc = geom.Point{c.WidthPt / 1.5, c.HeightPt / 1.5}
                case paint.Event:
                onPaint(c)
                a.EndPaint(e)
                case touch.Event:
                touchLoc = geom.Point{geom.Pt(e.X),geom.Pt(e.Y)}
            }
        }
    })
}

func onStart() {
    var err error
    program, err = glutil.CreateProgram(vertexShader, fragmentShader)
    if err != nil {
        log.Printf("error creating GL program: %v", err)
        return
    }

    //创建一个WebGLBuffer对象，把它绑定到顶点缓冲上，并把顶点数据载入到顶点冲。
    buf = gl.CreateBuffer()
    gl.BindBuffer(gl.ARRAY_BUFFER, buf)
    gl.BufferData(gl.ARRAY_BUFFER, triangleData, gl.STATIC_DRAW)

    /*opengl中三种变量
    uniform变量是外部application程序传递给（vertex和fragment）shader的变量。因此它是application通过函数glUniform**（）函数赋值的。
    在（vertex和fragment）shader程序内部，uniform变量就像是C语言里面的常量（const ），它不能被shader程序修改。（shader只能用，不能改）

    attribute变量是只能在vertex shader中使用的变量。（它不能在fragment shader中声明attribute变量，也不能被fragment shader中使用）
    一般用attribute变量来表示一些顶点的数据，如：顶点坐标，法线，纹理坐标，顶点颜色等。
    在application中，一般用函数glBindAttribLocation（）来绑定每个attribute变量的位置，然后用函数glVertexAttribPointer（）为每个attribute变量赋值。

    varying变量是vertex和fragment shader之间做数据传递用的。一般vertex shader修改varying变量的值，然后fragment shader使用该varying变量的值。
    因此varying变量在vertex和fragment shader二者之间的声明必须是一致的。application不能使用此变量。
    */
    position = gl.GetAttribLocation(program, "position")//获取位置对象(索引)
    color =  gl.GetUniformLocation(program, "color") // 获取颜色对象(索引)
    scan = gl.GetUniformLocation(program, "scan") // 获取偏移对象(索引)
    // fmt.Println(position.String(),color.String(),offset.String())//Attrib(0) Uniform(1) Uniform(0)
    // TODO(crawshaw): the debug package needs to put GL state init here
    // Can this be an event.Register call now??
}

//停止时触发,清理
func onStop() {
    gl.DeleteProgram(program)
    gl.DeleteBuffer(buf)
}


func onPaint(c size.Event) {
    //清场
    gl.ClearColor(1, 1, 1, 1) //设置背景颜色
    gl.Clear(gl.COLOR_BUFFER_BIT)

    //使用program
    gl.UseProgram(program)

    gl.Uniform4f(color, 0, 0.5, 0.8, 1)//设置color对象值,设置4个浮点数.
    //offset有两个值X,Y,窗口左上角为(0,0),右下角为(1,1)
    //gl.Uniform4f(offset,5.0,1.0,1.0,1.0 )
    //gl.Uniform2f(offset,offsetx,offsety )//为2参数的uniform变量赋值
    //log.Println("offset:",offsetx,offsety, 0, 0)
    gl.UniformMatrix4fv(scan, []float32{
        float32(touchLoc.X/c.WidthPt*4-2), 0, 0, 0,
        0, float32(touchLoc.Y/c.HeightPt*4-2), 0, 0,
        0, 0, 0, 0,
        0, 0, 0, 1,
    })
    gl.BindBuffer(gl.ARRAY_BUFFER, buf)
    gl.EnableVertexAttribArray(position)
    /*glVertexAttribPointer 指定了渲染时索引值为 index 的顶点属性数组的数据格式和位置。调用gl.vertexAttribPointer()方法，把顶点着色器中某个属性相对应的通用属性索引连接到绑定的webGLBUffer对象上。
    index 指定要修改的顶点属性的索引值
    size    指定每个顶点属性的组件数量。必须为1、2、3或者4。初始值为4。（如position是由3个（x,y,z）组成，而颜色是4个（r,g,b,a））
    type    指定数组中每个组件的数据类型。可用的符号常量有GL_BYTE, GL_UNSIGNED_BYTE, GL_SHORT,GL_UNSIGNED_SHORT, GL_FIXED, 和 GL_FLOAT，初始值为GL_FLOAT。
    normalized  指定当被访问时，固定点数据值是否应该被归一化（GL_TRUE）或者直接转换为固定点值（GL_FALSE）。
    stride  指定连续顶点属性之间的偏移量。如果为0，那么顶点属性会被理解为：它们是紧密排列在一起的。初始值为0。
    pointer 指定第一个组件在数组的第一个顶点属性中的偏移量。该数组与GL_ARRAY_BUFFER绑定，储存于缓冲区中。初始值为0；
    */
    gl.VertexAttribPointer(position, coordsPerVertex, gl.FLOAT, false, 0, 0) //更新position值
    gl.DrawArrays(gl.TRIANGLES, 0, vertexCount)
    gl.DisableVertexAttribArray(position)

    debug.DrawFPS(c)
}

var triangleData = f32.Bytes(binary.LittleEndian, //三角
0.0, 0.5, 0.0, // top left
-0.5, -0.5, 0.0, // bottom left
0.5, -0.5, 0.0, // bottom right
)

const (
    coordsPerVertex = 3 //点坐标
    vertexCount = 3 //点数
)

//两类着色器编程使用GLSL(GL Shader Language，GL着色语言)，它是OpenGL的一部分。与C或Java不同，GLSL必须在运行时编译，这意味着每次启动程序，所有的着色器将重新编译。
//顶点(vertex)着色器，它将作用于每个顶点上
//vec2即2个值,vec4即4个值
const vertexShader = `#version 100
uniform mat4 scan;
attribute vec4 position;
void main() {
	gl_Position = position*scan ;
}`

//片断（Fragment）着色器，它将作用于每一个采样点
const fragmentShader = `#version 100
precision mediump float;
uniform vec4 color;
void main() {
	gl_FragColor = color;
}`
