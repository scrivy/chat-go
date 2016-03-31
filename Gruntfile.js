'use strict';

module.exports = function(grunt) {
  require('load-grunt-tasks')(grunt);
  require('time-grunt')(grunt);

  grunt.initConfig({

    watch: {
      files: [
        'Gruntfile.js',
        'public/**/*'
      ],
      tasks: ['jshint'],
      options: {
        livereload: true
      }
    },

    jshint: {
      files: [
        'public/app.js',
        'public/components/**/*.js'
      ],
      options: {
        jshintrc: '.jshintrc'
      }
    }
  });
};
